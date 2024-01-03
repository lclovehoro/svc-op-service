package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"svc-op-service/db"
	"svc-op-service/jenkins"
	"svc-op-service/types"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

var JenkinsController = new(jenkinsController)

type jenkinsController struct {
}

func (j *jenkinsController) GetJenkinsJob(c *gin.Context) {
	var app types.App
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if app.Appid == "" {
		klog.V(3).ErrorS(fmt.Errorf("invalid arguments"), "Request.Body, the appId parameter is incorrect")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid arguments, the appId parameter is incorrect",
		})
		return
	}

	jjob, err := jenkins.JenkinsClient.GetJob(jenkins.JenkinsClient.Context, app.Appid)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    0,
				"message": err.Error(),
			},
		)
	}
	klog.Info(jjob.Raw)
}

func (j *jenkinsController) StartJob(c *gin.Context) {
	var release types.Release
	ctx := context.Background()

	if err := c.ShouldBindJSON(&release); err != nil {
		klog.V(3).ErrorS(err, "Failed to bind JSON", "types.App", release)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    0,
				"message": err.Error(),
			})
		return
	}

	if release.App.Releasetime != time.Now().Format("2006-01-02") {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "release time is not today",
			})
		return
	}

	// 查询是否在当日的发布单中
	rs, err := release.Search(db.MongoClient.Database)
	if err != nil {
		klog.V(3).ErrorS(err, "Error when search by name", "types.App", release)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    0,
				"message": err.Error(),
			})
		return
	}
	if rs == nil {
		klog.V(3).ErrorS(err, "Failed to search by name", "types.App", release)
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": fmt.Sprintf("not found appId %s", release.App.Appid),
			})
		return
	}

	if rs[0].Status.Prd.Packagestatus.Status == types.INQUEUE {
		klog.V(3).Info("don't click twice")
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": fmt.Sprintf("don't click twice, appId: %s", rs[0].App.Appid),
			})
		return
	}

	// 创建job对象
	job, err := jenkins.JenkinsClient.GetJob(ctx, release.App.Appid)
	klog.V(3).ErrorS(err, "get job failed", "jobName", release.App.Appid)
	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"code":    0,
				"message": err.Error(),
			})
		return
	}

	// 调用queue/item,开始build
	queueId, err := job.InvokeSimple(ctx, nil)
	klog.V(3).ErrorS(err, "get job queueid failed", "jobName", release.App.Appid)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    0,
				"message": err.Error(),
			})
		return
	}

	// 注入当前queueId以及queueId状态
	queueItem := types.JobQueue{
		Queueid: queueId,
		Status:  types.INQUEUE,
	}

	release.Status = rs[0].Status
	release.Status.Prd.Packagestatus = queueItem
	release.Status.Prd.PackageCounts++

	if err := release.UpdateStatus(db.MongoClient.Database); err != nil {
		klog.V(3).ErrorS(err, "Failed to update queue status", "release", release)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    0,
				"message": err.Error(),
			})
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":    0,
			"message": "build job success",
		})

	// 协程异步修改发布单打包状态
	go updatePackageStatus(ctx, &release, queueId)

}

func updatePackageStatus(ctx context.Context, release *types.Release, queueId int64) {
	build, err := jenkins.JenkinsClient.GetBuildFromQueueID(ctx, queueId)
	if err != nil {
		klog.V(3).ErrorS(err, "Failed to get build from queueid", "queueId", queueId)
		return
	}

	for build.IsRunning(ctx) {
		ticker := time.NewTicker(time.Second * 10)
		build.Poll(ctx)
		<-ticker.C
	}

	release.Status.Prd.Packagestatus = types.JobQueue{
		Queueid:     queueId,
		Buildnumber: build.GetBuildNumber(),
		Packagetime: build.GetTimestamp().Format("2006-01-02 15:04:05"),
		Status:      build.GetResult(),
	}
	release.Status.Prd.IsPackage = build.IsGood(ctx)

	if err := release.UpdateStatus(db.MongoClient.Database); err != nil {
		klog.V(3).ErrorS(err, "Failed to update release package status", "appid", release.App.Appid)
		return
	}

	klog.V(3).InfoS("update release package status success", "appid", release.App.Appid)
}

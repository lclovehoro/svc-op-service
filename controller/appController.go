package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"svc-op-service/db"
	"svc-op-service/types"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

var AppController = new(MyApp)

type MyApp struct {
}

// GET /path?releaseTime=yyyy-MM-DD&developers="developer1,developer2"&appId="appId1"'
func (a *MyApp) Get(c *gin.Context) {
	klog.V(3).InfoS("get query params", "params", c.Request.URL.Query())

	releaseTime := c.Query("releaseTime")
	if releaseTime == "" {
		releaseTime = time.Now().Format("2006-01-02")
	}

	if c.Query("appId") != "" {
		app := strings.Split(c.Query("appId"), ",")
		al, err := types.FindMany(db.MongoClient.Database, releaseTime, "app.appId", app...)
		if err != nil {
			klog.V(3).ErrorS(err, "something worry", "appId", app, "collection", releaseTime)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}
		if len(al) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    1000,
				"message": fmt.Sprintf("not found %s", app),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    al,
		})
		return
	}

	if c.Query("developers") != "" {
		developers := strings.Split(c.Query("developers"), ",")
		al, err := types.FindMany(db.MongoClient.Database, releaseTime, "app.developers", developers...)
		if err != nil {
			klog.V(3).ErrorS(err, "something worry", "developers", developers, "collection", releaseTime)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}
		if len(al) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    1000,
				"message": fmt.Sprintf("not found %s", developers),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    al,
		})
		return
	}

	if c.Query("appId") == "" && c.Query("developers") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid arguments, query parameters are required",
		})
		return
	}
}

// POST /path -d '{appId: "appA", releaseTime: "yyyy-MM-DD", mode: "jenkins", developers: "devUesr1, devUser2", dependApps: "APP1, APP2", desc: "xxxx", isSql: true}'
func (a *MyApp) Post(c *gin.Context) {
	var release types.Release
	if err := c.ShouldBindJSON(&release); err != nil {
		klog.V(3).ErrorS(err, "*gin.Context.ShouldBindJSON", "collection", release)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	klog.V(3).InfoS("Request.Body", "release", release)

	release.Status.Createtimestamp = time.Now().Format("2006-01-02 15:04:05")

	klog.V(3).InfoS("*App.POST.(NewApp)", "collection", release.App.Releasetime, "App", release)

	al, err := release.Search(db.MongoClient.Database)
	if err != nil {
		klog.V(3).ErrorS(err, "something worry for search", "collection", release.App.Releasetime, "Appid", release.App.Appid)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}
	if al != nil {
		klog.V(3).ErrorS(fmt.Errorf("%s已存在", release.App.Appid), fmt.Sprint(al))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    0,
			"message": fmt.Sprintf("%s已存在", release.App.Appid),
		})
		return
	}

	if err := release.Add(db.MongoClient.Database); err != nil {
		klog.V(3).ErrorS(err, "something worry for add", "collection", release.App.Releasetime, "App", release)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    release,
	})
}

// DELETE /path -d '{appId : "appId" , releaseTime: "yyyy-MM-DD"}'
func (a *MyApp) Delete(c *gin.Context) {
	var release types.Release

	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	result, err := release.Delete(db.MongoClient.Database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}
	if result == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    0,
			"message": fmt.Sprintf("not found %s", release.App.Appid),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": fmt.Sprintf("delete %s success", release.App.Appid),
	})
}

// DELEDE /path -d '{appId : ["app1","app2"], releaseTime: "yyyy-MM-DD"}'
func (a *MyApp) DeleteBatch(c *gin.Context) {
	var reqData types.DelBatchReqData

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if reqData.Releasetime == "" {
		reqData.Releasetime = time.Now().Format("2006-01-02")
	}

	result, err := reqData.DeleteBatch(db.MongoClient.Database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}
	if result == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("not found %s", reqData.Appid),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": fmt.Sprintf("delete %s success", reqData.Appid),
	})

}

// UPDATE /path -d '{appId: "appA", releaseTime: "yyyy-MM-DD", mode: "jenkins", developers: "devUesr1, devUser2", dependApps: "APP1, APP2", desc: "xxxx", isSql: true}'
func (a *MyApp) Put(c *gin.Context) {
	var release types.Release
	if err := c.ShouldBindJSON(&release); err != nil {
		klog.V(3).ErrorS(err, "Request.Body, the appId parameter is incorrect")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	rs, err := release.Search(db.MongoClient.Database)
	if err != nil {
		klog.V(3).ErrorS(err, "search failed, something has worry", "collection", release.App.Releasetime, "appId", release.App.Appid)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": err.Error(),
		})
	}

	klog.V(3).InfoS("search result", "collection", release.App.Releasetime, "release", rs)
	if rs == nil {
		klog.V(3).InfoS("no update is needed because it was not retrieved", "collection", release.App.Releasetime, "appId", release.App.Appid)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    0,
			"message": fmt.Sprintf("no update is needed because it was not retrieved, collection: %s, appId: %s", release.App.Releasetime, release.App.Appid),
		})
		return
	}

	for _, v := range rs {
		v.Status.Dev.Updatestatus = append(v.Status.Dev.Updatestatus, types.UpdateStatus{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Histroy:   v.App,
		})
		release.Status.Dev.Updatestatus = v.Status.Dev.Updatestatus
		klog.V(3).Infof("new release will update: %v", release)
		if err := release.Update(db.MongoClient.Database); err != nil {
			klog.V(3).ErrorS(err, "release update has something worry", "collection", release.App.Releasetime, "App", release.App)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    release.App,
	})
}

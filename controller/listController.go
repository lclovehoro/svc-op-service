package controller

import (
	"fmt"
	"net/http"
	"strings"
	"svc-op-service/db"
	"svc-op-service/types"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

var AppListsController = new(MyReleaseLists)

type MyReleaseLists struct {
}

// GET /path?releaseTime=xxxx.xx.xx&appId=xxx
func (al *MyReleaseLists) Get(c *gin.Context) {
	klog.V(3).InfoS("get query params", "params", c.Request.URL.Query())
	var releases types.Releases
	releaseTime, ok := c.GetQuery("releaseTime")
	if !ok {
		releaseTime = time.Now().Format("2006-01-02")
	}
	if releaseTime == "" {
		releaseTime = time.Now().Format("2006-01-02")
	}

	if err := releases.RelasesList(db.MongoClient.Database, releaseTime, strings.Split(c.Query("appId"), ",")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	if len(releases.List) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    1000,
			"message": fmt.Sprintf("no data was found on %s", releaseTime),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    releases,
	})
}

func (al *MyReleaseLists) Post(c *gin.Context) {

}

func (al *MyReleaseLists) Delete(c *gin.Context) {

}

func (al *MyReleaseLists) Put(c *gin.Context) {

}

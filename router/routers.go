package router

import (
	"svc-op-service/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	config := cors.Config{
		AllowAllOrigins: true,
		//AllowOrigins:    []string{"*"},
		AllowHeaders: []string{"Origin"},
	}
	r.Use(cors.New(config))
	// 注册路由
	v1 := r.Group("/api/v2")
	{
		// AUTH
		v1.GET("login/code", controller.LoginController.GetCode)
		v1.GET("users/info", controller.LoginController.GetUserInfo)
		v1.POST("users/login", controller.LoginController.GetToken)

		// RELEASE
		v1.GET("/release/get", controller.AppController.Get)
		v1.DELETE("/release/delete", controller.AppController.Delete)
		v1.DELETE("/release/delbatch", controller.AppController.DeleteBatch)
		v1.PUT("/release/update", controller.AppController.Put)
		v1.POST("/release/create", controller.AppController.Post)

		v1.GET("/release/list", controller.AppListsController.Get)

		// JENKINS JOB
		v1.GET("/job/get", controller.JenkinsController.GetJenkinsJob)
		v1.POST("/job/build/start", controller.JenkinsController.StartJob)
	}
	return r
}

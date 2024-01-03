package controller

import (
	"net/http"

	"svc-op-service/types"

	"github.com/gin-gonic/gin"
)

var LoginController = &MyLogin{}

type MyLogin struct {
}

func (l *MyLogin) GetCode(c *gin.Context) {
	// 获取验证码
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "OK",
		"data":    "123456",
	})
}

func (l *MyLogin) GetToken(c *gin.Context) {
	token := types.NewToken()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "OK",
		"data":    token,
	})
}

func (l *MyLogin) GetUserInfo(c *gin.Context) {
	//types.NewUser()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "OK",
		"data": map[string]interface{}{
			"username": "张三",
			"role":     []string{},
		},
	})
}

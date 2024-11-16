package server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RunHTTPServer(serviceName string, wrapper func(router *gin.Engine)) {
	//配置化
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		panic(errors.New("addr is empty"))
	}
	RunHTTPServerOnAddr(addr, wrapper)
}

func RunHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) {
	//使用匿名函数wrapper来实现不同的服务使用同一份代码
	apiRouter := gin.New()
	wrapper(apiRouter)
	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}

package main

import (
	"IM-system/internal/config"
	"IM-system/internal/httpserver"
	"IM-system/internal/logger"
	"IM-system/server"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger.Init()
	cfg := config.Load()
	//创建并初始化一个 Server 对象。
	s := server.NewServer(cfg.TCP.Host, cfg.TCP.Port)
	logger.Log.Info(
		"tcp server starting",
		"host", cfg.TCP.Host,
		"port", cfg.TCP.Port,
	)
	// TCP服务
	go s.Start()

	// gin服务
	//r := gin.Default()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(httpserver.Recovery())
	r.Use(httpserver.RequestLogger())
	
	httpserver.RegisterRoutes(r, s)//依赖传递
	//静态资源
	r.Static("/web", "./web") //浏览器访问：/web/xxxx去项目中的：./web/xxxx找文件
	HttpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	logger.Log.Info(
		"http server starting",
		"addr", HttpAddr,
	)
	go r.Run(HttpAddr)
	quit := make(chan os.Signal, 1) //存信号
	//告诉 Go runtime：如果收到 SIGINT/SIGTERM，就把这个信号写入 quit channel。
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("server shutting down")
	s.Shutdown()
}

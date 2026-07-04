package main

import (
	"IM-system/internal/config"
	"IM-system/internal/logger"
	"IM-system/server"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	logger.Init()
	cfg := config.Load()
	s := server.NewServer(cfg.TCP.Host, cfg.TCP.Port)
	logger.Log.Info(
		"tcp server starting",
		"host", cfg.TCP.Host,
		"port", cfg.TCP.Port,
	)
	// TCP服务
	go s.Start()

	// gin服务
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "pong",
		})
	})

	r.GET("/OnlineUsers", func(c *gin.Context) {
		users := s.GetOnlineUsers()
		c.JSON(http.StatusOK, gin.H{
			"OnlineUsers": users,
		})
	})
	r.GET("/Rooms", func(c *gin.Context) {
		rooms := s.GetRooms()
		c.JSON(http.StatusOK, gin.H{
			"Rooms": rooms,
		})
	})
	r.GET("/rooms/:room/members", func(c *gin.Context) {
		room := strings.TrimSpace(c.Param("room")) //直接拿干净数据
		if room == "" {                            //清理空格 防止脏数据
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid room parameter",
			})
			return
		}
		members, ok := s.GetMembers(room)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "room not found",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"members": members,
		})
	})
	r.POST("/rooms/:room/join", func(c *gin.Context) {
		room := strings.TrimSpace(c.Param("room"))

		if room == "" { //清理空格 防止脏数据
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid room parameter",
			})
			return
		}
		user := strings.TrimSpace(c.Query("user"))
		if user == "" { //清理空格 防止脏数据
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user parameter",
			})
			return
		}
		msg, ok := s.JoinRoom1(user, room)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": msg,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
	})
	r.POST("/leave", func(c *gin.Context) {
		user := strings.TrimSpace(c.Query("user"))
		if user == "" { //清理空格 防止脏数据
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user parameter",
			})
			return
		}
		msg, ok := s.LeaveRoom1(user)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": msg,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
	})
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

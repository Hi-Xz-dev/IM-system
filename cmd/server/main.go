package main

import (
	"IM-system/server"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func main() {
	s := server.NewServer("127.0.0.1", 8080)

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
	r.Run(":8081")
}

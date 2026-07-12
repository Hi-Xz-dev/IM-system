package httpserver

import(
	"github.com/gin-gonic/gin"
	"IM-system/server"

)
func RegisterRoutes(r *gin.Engine, s *server.Server){
	h := NewHandler(s)
	//r.Use(RequestLogger())//配置整个 Gin Engine
	r.GET("/ping", h.Ping)//注册路由
	r.GET("/rooms", h.Rooms)
	r.GET("/users/:user/rooms", h.UserRooms)
	r.GET("/online-users", h.OnlineUsers)
	r.POST("/rooms", h.CreateRoom)
	r.PUT("/user/:user", h.Rename)
	r.GET("/rooms/:room/members", h.Members)
	r.POST("/rooms/:room/members/:user", h.Join)
	r.DELETE("/rooms/:room/members/:user", h.Leave)	
}
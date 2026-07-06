package httpserver

import(
	"github.com/gin-gonic/gin"
	"IM-system/server"

)
func RegisterRoutes(r *gin.Engine, s *server.Server){
	h := NewHandler(s)
	r.Use(RequestLogger())
	r.GET("/ping", h.Ping)
	r.GET("/online-users", h.OnlineUsers)
	r.GET("/rooms", h.Rooms)
	r.GET("/rooms/:room/members", h.Members)
	r.POST("/rooms/:room/join", h.Join)
	r.POST("/leave", h.Leave)	
}
package httpserver

import (
	"IM-system/gateway/ws"
	"IM-system/internal/auth"
	"IM-system/internal/middleware"
	"IM-system/server"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	s *server.Server,
	authService *auth.Service,
	gateway *ws.Gateway,
	authMiddleware *middleware.AuthMiddleware,
) {
	h := NewHandler(s, authService)
	//r.Use(RequestLogger())//配置整个 Gin Engine
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.GET("/ping", h.Ping) //注册路由

	r.GET("/ws", gateway.Handler)

	//需要登录接口
	api := r.Group("/api")

	api.Use(
		authMiddleware.Handler(),
	)
	api.GET("/rooms", h.Rooms)
	api.GET("/users/:user/rooms", h.UserRooms)
	api.GET("/online-users", h.OnlineUsers)

	api.POST("/rooms", h.CreateRoom)
	api.PUT("/user/:user", h.Rename)
	api.GET("/rooms/:room/members", h.Members)
	api.POST("/rooms/:room/members/:user", h.Join)
	api.DELETE("/rooms/:room/members/:user", h.Leave)

}

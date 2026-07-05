package httpserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"IM-system/server"
)

type Handler struct {
	server *server.Server//保存了一个 Server 指针
}
//构造函数。把 main.go 里的 s 注入到 Handler 里面
//这一步叫：依赖注入 Dependency Injection
func NewHandler(s *server.Server) *Handler {
	return &Handler{server : s,
	}
}

func (h *Handler)Ping(c *gin.Context) {
	c.JSON(http.StatusOK, OK("pong"))
}

func (h *Handler)OnlineUsers(c *gin.Context) {
	users := h.server.GetOnlineUsers()
	c.JSON(http.StatusOK, OK(users))
}

func (h *Handler)Rooms(c *gin.Context) {
	rooms := h.server.GetRooms()
	c.JSON(http.StatusOK,OK(rooms))
}

func (h *Handler)Members(c *gin.Context) {
	room := strings.TrimSpace(c.Param("room")) //直接拿干净数据
	if room == "" {                            //清理空格 防止脏数据
		c.JSON(http.StatusBadRequest, Fail("invalid room parameter"))
		return
	}
	members, ok := h.server.GetMembers(room)
	if !ok {
		c.JSON(http.StatusNotFound,Fail("room not found"))
		return
	}
	c.JSON(http.StatusOK, OK(members))
}
func (h *Handler)Join(c *gin.Context) {
	room := strings.TrimSpace(c.Param("room"))

	if room == "" { //清理空格 防止脏数据
		c.JSON(http.StatusBadRequest, Fail("invalid room parameter"))
		return
	}
	user := strings.TrimSpace(c.Query("user"))
	if user == "" { //清理空格 防止脏数据
		c.JSON(http.StatusBadRequest, Fail("invalid user parameter"))
		return
	}
	msg, ok := h.server.JoinRoom1(user, room)
	if !ok {
		c.JSON(http.StatusNotFound, Fail(msg))
		return
	}
	c.JSON(http.StatusOK, OK(msg))
}

func (h *Handler)Leave(c *gin.Context) {
	user := strings.TrimSpace(c.Query("user"))
	if user == "" { //清理空格 防止脏数据
		c.JSON(http.StatusBadRequest, Fail("invalid user parameter"))
		return
	}
	msg, ok := h.server.LeaveRoom1(user)
	if !ok {
		c.JSON(http.StatusNotFound, Fail(msg))
		return
	}
	c.JSON(http.StatusOK, OK(msg))
}

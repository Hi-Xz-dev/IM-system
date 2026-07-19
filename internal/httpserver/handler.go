package httpserver

import (
	"IM-system/internal/service"
	"IM-system/server"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"IM-system/internal/auth"
)

type Handler struct {
	server      *server.Server //保存了一个 Server 指针
	roomService *service.RoomService
	authService *auth.Service
}

// 构造函数 把 main.go 里的 s 注入到 Handler 里面
// 这一步叫：依赖注入 Dependency Injection
func NewHandler(s *server.Server,authService *auth.Service,) *Handler {
	return &Handler{
		server:      s,
		roomService: service.NewRoomService(s),
		authService: authService,
	}
}

func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, OK("pong"))
}

func (h *Handler) OnlineUsers(c *gin.Context) {
	users := h.server.GetOnlineUsers()
	c.JSON(http.StatusOK, OK(users))
}

func (h *Handler) Rooms(c *gin.Context) {
	rooms := h.roomService.GetRooms()
	c.JSON(http.StatusOK, OK(rooms))
}

// 当前成员
func (h *Handler) Members(c *gin.Context) {
	room, ok := getRoomParam(c)
	if !ok {
		return
	}
	members, ok := h.server.GetMembers(room)
	if !ok {
		c.JSON(http.StatusNotFound, Fail("room not found"))
		return
	}
	c.JSON(http.StatusOK, OK(members))
}

// 加入房间
func (h *Handler) Join(c *gin.Context) {
	room, ok := getRoomParam(c)
	if !ok {
		return
	}
	user, ok := getUserParam(c)
	if !ok {
		return
	}
	msg, ok := h.server.JoinRoomByName(user, room)
	if !ok {
		c.JSON(http.StatusNotFound, Fail(msg))
		return
	}
	c.JSON(http.StatusOK, OK(msg))
}

// 离开房间
func (h *Handler) Leave(c *gin.Context) {
	room, ok := getRoomParam(c)
	if !ok {
		return
	}
	user, ok := getUserParam(c)
	if !ok {
		return
	}
	msg, ok := h.server.LeaveRoomByName(user, room)
	if !ok {
		c.JSON(http.StatusNotFound, Fail(msg))
		return
	}
	c.JSON(http.StatusOK, OK(msg))
}

// 定义请求DTO
type RenameUserRequest struct {
	Name string `json:"name"`
}

// 用户改名
func (h *Handler) Rename(c *gin.Context) {
	oldName, ok := getUserParam(c)
	if !ok {
		return
	}
	var req RenameUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid new user name"))
		return
	}
	msg, success := h.server.RenameByName(oldName, req.Name)
	if !success {
		c.JSON(http.StatusNotFound, Fail(msg))
	}
	c.JSON(http.StatusOK, OK(msg))
}

func (h *Handler) UserRooms(c *gin.Context) {
	userName, ok := getUserParam(c)
	if !ok {
		return
	}
	rooms, found := h.server.GetUserRooms(userName)
	if !found {
		c.JSON(http.StatusNotFound, Fail("user not found"))
		return
	}
	c.JSON(http.StatusOK, OK(rooms))
}

type CreateRoomRequest struct {
	Name string `json:"name"`
	User string `json:"user"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Fail("invaild request body"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.User = strings.TrimSpace(req.User)

	if req.User == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid user name"))
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid room name"))
		return
	}
	msg, ok := h.server.CreateRoomByName(req.User, req.Name)
	if !ok {
		c.JSON(http.StatusConflict, Fail(msg))
		return
	}
	c.JSON(http.StatusCreated, OK(msg))
}

package httpserver

import (
	"IM-system/internal/auth"
	"IM-system/internal/logger"
	"IM-system/internal/message"
	"IM-system/internal/repository"
	"IM-system/internal/service"
	"IM-system/server"

	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	server         *server.Server //保存了一个 Server 指针
	roomService    *service.RoomService
	authService    *auth.Service
	messageService *message.MessageService
}

// 构造函数 把 main.go 里的 s 注入到 Handler 里面
// 这一步叫：依赖注入 Dependency Injection
func NewHandler(
	s *server.Server,
	authService *auth.Service,
	roomRepo *repository.RoomRepository,
	messageService *message.MessageService,
) *Handler {
	return &Handler{
		server:         s,
		roomService:    service.NewRoomService(s, roomRepo),
		authService:    authService,
		messageService: messageService,
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

// 获取房间历史信息
func (h *Handler) RoomHistory(c *gin.Context) {
	roomID, ok := getRoomParam(c)

	if !ok {
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
		return
	}
	before, err := message.ParseHistoryCursor(c.Query("before"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Fail(err.Error()))
		return
	}
	page, err := h.server.GetRoomHistoryForUser(
		c.Request.Context(),
		userID,
		roomID,
		before,
	)

	if writeRoomAccessError(c, err, "get room history failed") {
		return
	}

	writeHistoryPage(c, page)
}

// 获取用户历史信息
func (h *Handler) PrivateHistory(c *gin.Context) {
	userID, ok := getUserID(c)

	if !ok {
		c.JSON(
			http.StatusUnauthorized,
			Fail("not login"),
		)
		return
	}

	targetID, err := strconv.ParseInt(
		c.Param("user"),
		10,
		64,
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			Fail("invalid user id"),
		)
		return
	}

	before, err := message.ParseHistoryCursor(c.Query("before"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Fail(err.Error()))
		return
	}
	page, err := h.messageService.GetPrivateHistory(
		c.Request.Context(),
		userID,
		targetID,
		before,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			Fail("get user history failed"),
		)
		return
	}

	writeHistoryPage(c, page)
}

func writeHistoryPage(c *gin.Context, page message.HistoryPage) {
	if page.NextCursor != "" {
		c.Header("X-Next-Cursor", page.NextCursor)
	}
	c.JSON(http.StatusOK, OK(page.Messages))
}

// 当前成员
func (h *Handler) Members(c *gin.Context) {
	room, ok := getRoomParam(c)
	if !ok {
		return
	}
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
		return
	}
	members, err := h.server.GetMembers(userID, room)
	if writeRoomAccessError(c, err, "get room members failed") {
		return
	}
	c.JSON(http.StatusOK, OK(members))
}

func writeRoomAccessError(c *gin.Context, err error, fallback string) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, server.ErrRoomNotFound):
		c.JSON(http.StatusNotFound, Fail("room not found"))
	case errors.Is(err, server.ErrNotRoomMember):
		c.JSON(http.StatusForbidden, Fail("join room before viewing its contents"))
	default:
		c.JSON(http.StatusInternalServerError, Fail(fallback))
	}
	return true
}

// 加入房间
func (h *Handler) Join(c *gin.Context) {
	roomID, ok := getRoomParam(c)
	if !ok {
		return
	}

	userID, ok := getUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
		return
	}

	msg, ok := h.server.JoinRoomH(userID, roomID)
	if !ok {
		c.JSON(http.StatusNotFound, Fail(msg))
		return
	}
	c.JSON(http.StatusOK, OK(msg))
}

// 离开房间
func (h *Handler) Leave(c *gin.Context) {
	roomID, ok := getRoomParam(c)
	if !ok {
		return
	}

	userID, ok := getUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
		return
	}

	msg, ok := h.server.LeaveRoomH(userID, roomID)
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

	userID, ok := getUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
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
	err := h.server.RenameH(c.Request.Context(), userID, req.Name)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, OK("修改用户名成功"))
	case errors.Is(err, auth.ErrInvalidNickname):
		c.JSON(http.StatusBadRequest, Fail("昵称须为 1 到 64 个字符"))
	case errors.Is(err, server.ErrRenameUserOffline):
		c.JSON(http.StatusNotFound, Fail("未找到在线用户"))
	default:
		logger.Log.Error("rename user failed", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, Fail("修改昵称失败，请稍后重试"))
	}
}

func (h *Handler) UserRooms(c *gin.Context) {

	userID, ok := getUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, Fail("not login"))
		return
	}

	rooms, found := h.server.GetUserRoomsH(userID)

	if !found {
		c.JSON(http.StatusNotFound, Fail("user not found"))
		return
	}
	c.JSON(http.StatusOK, OK(rooms))
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Fail("invaild request body"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid room name"))
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(
			http.StatusUnauthorized,
			Fail("not login"),
		)
		return
	}

	_, err := h.roomService.CreateRoom(
		c.Request.Context(),
		userID,
		req.Name,
	)
	if err != nil {
		if errors.Is(err, repository.ErrRoomAlreadyExists) {
			c.JSON(
				http.StatusConflict,
				Fail("room already exists"),
			)
			return
		}

		logger.Log.Error(
			"create room record failed",
			"user_id", userID,
			"room_name", req.Name,
			"error", err,
		)

		c.JSON(
			http.StatusInternalServerError,
			Fail("create room failed"),
		)
		return
	}

	c.JSON(http.StatusCreated, OK("创建房间成功"))
}

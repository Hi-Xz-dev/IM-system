package server

import (
	"IM-system/internal/auth"
	"IM-system/room"
	"IM-system/user"
	"net"
	"sync"
)

type Server struct {
	//服务器监听地址
	IP   string
	Port int
	//在线用户列表
	OnlineUsers map[int64][]*user.User
	//保护一组有关联的共享状态
	mapLock sync.RWMutex
	//消息广播的channel
	Message chan string
	//增加房间列表
	Rooms map[string]*room.Room
	//断开用户链接
	Disconnect chan *user.User
	//服务端的 TCP 监听器
	listener net.Listener
	//关闭
	IsShutdown  bool
	authService *auth.Service
}

// 创建一个server接口
func NewServer(ip string, port int, authService *auth.Service) *Server {
	return &Server{
		IP:          ip,
		Port:        port,
		OnlineUsers: make(map[int64][]*user.User),
		Message:     make(chan string, 100),
		Rooms:       make(map[string]*room.Room),
		Disconnect:  make(chan *user.User, 100),
		authService: authService,
	}
}

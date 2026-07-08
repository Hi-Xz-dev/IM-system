package server

import (
	"IM-system/room"
	"IM-system/user"
	"net"
	"sync"
)

type Server struct {
	IP   string
	Port int
	//在线用户列表
	OnlineUsers map[string]*user.User
	mapLock     sync.RWMutex //防止多人篡改数据
	//消息广播的channel
	Message chan string
	//增加房间列表
	Rooms map[string]*room.Room
	//断开用户链接
	Disconnect chan *user.User
	//退出
	listener net.Listener
	//
	IsShutdown bool
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:   ip,
		Port: port,
		//用户信息
		OnlineUsers: make(map[string]*user.User),
		Message:     make(chan string, 100),
		//房间信息
		Rooms:      make(map[string]*room.Room),
		Disconnect: make(chan *user.User, 100),
	}
}


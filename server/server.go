package server

import (
	"IM-system/room"
	"IM-system/user"
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
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

// 创建监听调用链接
func (s *Server) ListenDisconnect() {
	for {
		usr := <-s.Disconnect
		s.Offline(usr)
	}

}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User 非阻塞广播
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		//将msg发送给全部的在线User
		s.mapLock.RLock()
		users := make([]*user.User, 0, len(s.OnlineUsers))
		for _, cli := range s.OnlineUsers {
			users = append(users, cli)
		}
		s.mapLock.RUnlock()
		for _, cli := range users {
			cli.SendMsg(msg)
		}
	}
}

// 超时强踢功能
func (s *Server) CleanOnlineUser() {
	for {
		var kickList []*user.User
		s.mapLock.RLock()

		for _, u := range s.OnlineUsers {
			if u.ActiveTime != 0 &&
				time.Now().Unix()-u.ActiveTime > 6000 {
				kickList = append(kickList, u)
			}
		}
		s.mapLock.RUnlock()
		for _, u := range kickList {
			s.Offline(u)
		}
		time.Sleep(10 * time.Second)
	}
}

// 广播消息方法
func (s *Server) BroadCast(user *user.User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// 房间广播纯IO
func (s *Server) RoomBroadcast(users []*user.User, msg string) {
	for _, u := range users {
		u.SendMsg(msg)
	}
}

func (s *Server) Shutdown() {
	s.IsShutdown = true
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.mapLock.RLock()
	users := make([]*user.User, 0, len(s.OnlineUsers))
	for _, u := range s.OnlineUsers {
		users = append(users, u)
	}
	s.mapLock.RUnlock()
	for _, u := range users {
		u.SendMsg("[系统] 服务器正在关闭")
		s.Offline(u)
	}

}

// handler
func (s *Server) Handler(conn net.Conn) {
	//..当前链接的业务
	usr := user.NewUser(conn)
	go usr.ListenMessage(s.Disconnect)
	//通知 Handler：读协程结束了，你也可以退出了 只做通知 内容无所谓 省内存
	done := make(chan struct{})
	//用户上线业务
	s.Online(usr)
	//启动读协程 负责读客户端发来的消息
	go func() {
		defer close(done)//通知 Handler：读协程已退出
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() { //返回bool
			usr.UpdateActiveTime()
			msg := strings.TrimSpace(scanner.Text()) //取出读到的内容
			if msg == "" {
				continue
			}

			if msg == "quit" {
				s.Offline(usr)
				return
			}
			//用户针对msg进行消息处理
			s.DoMessage(usr, msg)
		}
		if err := scanner.Err(); err != nil { //错误可能有两种 一种正常，另外是真错了
			fmt.Println("Conn Read err:", err)
		}
		s.Offline(usr)//业务

	}()
	<-done//等这个 Channel 被关闭。
}

// 启动服务器接口
func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	s.listener = listener
	//close listen socket
	defer listener.Close()
	//启动超时踢人功能
	go s.CleanOnlineUser()
	go s.ListenDisconnect()
	//启动监听Message的goroutine
	go s.ListenMessager()
	for {
		//accept等待 TCP三次握手
		conn, err := listener.Accept()
		if err != nil {
			if s.IsShutdown {
				fmt.Println("server shutdown, stop accepting connections")
				return
			}
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go s.Handler(conn) //goroutine
	}
}

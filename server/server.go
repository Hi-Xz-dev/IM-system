package server

import (
	"IM-system/room"
	"IM-system/user"
	"fmt"
	"io"
	"log"
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
		Rooms: make(map[string]*room.Room),
		Disconnect: make(chan *user.User,100),
	}
}
//创建监听调用链接
func (s *Server) ListenDisconnect(){
	for{
	usr := <- s.Disconnect
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
			select {
			case cli.C <- msg:
				// 消息发送成功
			default:
				log.Printf("drop message to user %s: channel full", cli.Name)

			}
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
		select {
		case u.C <- msg:
			 // 消息发送成功
		default: log.Printf("drop message to user %s: channel full", u.Name)
		}
	}
}

// handler
func (s *Server) Handler(conn net.Conn) {
	//..当前链接的业务
	//fmt.Println("链接建立成功")
	usr := user.NewUser(conn)
	go usr.ListenMessage(s.Disconnect)
	//用户上线业务

	s.Online(usr)
	//接受客户端发送的信息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			usr.UpdateActiveTime()
			if n == 0 {
				s.Offline(usr)
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				s.Offline(usr)
				conn.Close()
				return
			}
			//提取用户的信息
			msg := strings.TrimSpace(string(buf[:n]))
			if msg == "quit" {
				s.Offline(usr)
				conn.Close()
				return
			}
			//用户针对msg进行消息处理
			s.DoMessage(usr, msg)
		}
	}()
	//当前handler阻塞
	select {}
}

// 启动服务器接口
func (s *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listen.Close()
	//启动超时踢人功能
	go s.CleanOnlineUser()
	go s.ListenDisconnect()
	//启动监听Message的goroutine
	go s.ListenMessager()
	for {
		//accept等待 TCP三次握手
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go s.Handler(conn) //goroutine
	}
}

package user

import (
	"fmt"
	"net"
	"time"
)

type User struct {
	Name string      //昵称
	Addr string      //地址
	C    chan string //消息队列
	conn net.Conn    //TCP连接

	ActiveTime int64 //心跳

	CurrentRoom string //当前房间
}

// 创建一个用户API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	u := &User{
		Name:       userAddr,
		Addr:       userAddr,
		C:          make(chan string, 100),
		conn:       conn,
		ActiveTime: time.Now().Unix(), //时间初始化
	}
	//go u.ListenMessage(s.Disconnect)
	//启动监听当前user channel消息的goroutine
	return u
}

// 更新活跃时间
func (u *User) UpdateActiveTime() {
	u.ActiveTime = time.Now().Unix()
}

// 给当前user对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	select {
	case u.C <- msg:
	default:
		fmt.Println("[WARN] user channel full, drop message:", u.Name)
	}
}

// 持续监听 User.C，并将消息发送给客户端。
// 若写超时或写失败，则通知 Server 统一处理连接断开。
// 这个函数只能往 channel 里发送，不能接收。
func (u *User) ListenMessage(disconnect chan<- *User) {
	for msg := range u.C {
		deadline := time.Now().Add(5 * time.Second)
		if err := u.conn.SetWriteDeadline(deadline); err != nil {
			select {
			case disconnect <- u:
			default:
			}
			return
		}
		if _, err := u.conn.Write([]byte(msg + "\n")); err != nil {
			select {
			case disconnect <- u:
			default:
			}
			return
		}
	}
}

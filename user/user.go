package user

import (
	"fmt"
	"net"
	"time"
)

type User struct {
	Nickname string      //昵称
	ID       int64       //数据库用户ID
	Addr     string      //TCP地址
	C        chan string //消息队列
	conn     net.Conn    //TCP连接

	ActiveTime int64 //心跳

	JoinedRooms map[string]struct{} //用户加入哪些房间

	IsClosed bool
}

// 创建一个用户API
func NewUser(conn net.Conn, id int64, nickname string,) *User {
	userAddr := conn.RemoteAddr().String()

	u := &User{
		Nickname:    nickname,
		ID:          id,
		Addr:        userAddr,
		C:           make(chan string, 100),
		conn:        conn,
		ActiveTime:  time.Now().Unix(),
		JoinedRooms: make(map[string]struct{}),
	}
	return u
}

func (u *User) Close() {
	if u.conn != nil {
		_ = u.conn.Close()
	}
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
		fmt.Println("[WARN] user channel full, drop message:", u.Nickname)
	}
}

// 监听某一个用户的信息
// 持续监听 User.C，并将消息发送给客户端。
// 若写超时或写失败，则通知 Server 统一处理连接断开。
// 这个函数只能往 channel 里发送，不能接收。
func (u *User) ListenMessage(disconnect chan<- *User) {
	for msg := range u.C {
		deadline := time.Now().Add(5 * time.Second)
		if err := u.conn.SetWriteDeadline(deadline); err != nil { //本身失败直接返回 说明底层存在问题 等待下面TCP写入 最多5秒
			select {
			case disconnect <- u:
			default:
			}
			return
		}
		if _, err := u.conn.Write([]byte(msg + "\n")); err != nil { //TCP写入数据
			select {
			case disconnect <- u:
			default:
			}
			return
		}
	}
}

func (u *User) InRoom(roomName string) bool {
	_, ok := u.JoinedRooms[roomName]
	return ok
}

func (u *User) AddRoom(roomName string) {
	u.JoinedRooms[roomName] = struct{}{}
}

func (u *User) RemoveRoom(roomName string) {
	delete(u.JoinedRooms, roomName)
}

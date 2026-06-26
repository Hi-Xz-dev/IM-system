package user
import (
	"net"
	"time"
)

type User struct{
	Name string//昵称
	Addr string//地址
	C    chan string//消息队列
	conn net.Conn//TCP连接

	ActiveTime int64//心跳

	CurrentRoom string//当前房间
}

//创建一个用户API
func NewUser(conn net.Conn) *User{
	userAddr := conn.RemoteAddr().String()

	u := &User{
		Name:userAddr,
		Addr:userAddr,
		C:make(chan string),
		conn:conn,
		ActiveTime: time.Now().Unix(),//时间初始化
	}
	go u.ListenMessage()
	//启动监听当前user channel消息的goroutine
	return u
}

//更新活跃时间
func (u *User)UpdataActiveTime(){
	u.ActiveTime = time.Now().Unix()
}

//给当前user对应的客户端发送消息
func (u *User) SendMsg(msg string){
	u.conn.Write([]byte(msg))
}

//监听当前User channel 方法，有消息 直接发送给客户端
func (u *User) ListenMessage(){
	for msg := range u.C{
		u.conn.Write([]byte(msg+"\n"))
	}
}


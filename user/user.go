package user

import (
	"IM-system/internal/connection"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrSendQueueFull    = errors.New("send queue full")
	ErrConnectionClosed = errors.New("connection closed")
)

const defaultQueueFullTimeout = 5 * time.Second

type User struct {
	Nickname string      //昵称
	ID       int64       //数据库用户ID
	Addr     string      //TCP地址
	C        chan string //消息队列
	conn     connection.Connection

	activeTime atomic.Int64 // 活跃时间通过方法进行原子读写

	JoinedRooms map[int64]struct{} //用户加入哪些房间

	IsClosed bool

	// 发送生命周期独立于 Server 的成员状态锁；不关闭 C，避免并发发送 panic。
	sendMu           sync.Mutex
	done             chan struct{}
	closed           bool
	congestedSince   time.Time
	congestionTimer  *time.Timer
	queueFullTimeout time.Duration
}

// 创建一个用户API
func NewUser(conn connection.Connection, id int64, nickname string, addr string) *User {

	u := &User{
		Nickname:    nickname,
		ID:          id,
		Addr:        addr,
		C:           make(chan string, 100),
		conn:        conn,
		JoinedRooms: make(map[int64]struct{}),
	}
	u.UpdateActiveTime()
	return u
}

func (u *User) Close() {
	u.sendMu.Lock()
	if u.closed {
		u.sendMu.Unlock()
		return
	}
	u.closeSendLocked()
	u.sendMu.Unlock()
	if u.conn != nil {
		_ = u.conn.Close()
	}
}

func (u *User) doneLocked() chan struct{} {
	if u.done == nil {
		u.done = make(chan struct{})
	}
	return u.done
}

func (u *User) closeSendLocked() {
	u.closed = true
	close(u.doneLocked())
	if u.congestionTimer != nil {
		u.congestionTimer.Stop()
	}
}

// 第一次队列满 → 返回错误，并开始 5 秒计时
// 之后继续满 → 返回错误，不重复创建或延长计时器
func (u *User) startCongestionTimerLocked() {
	if !u.congestedSince.IsZero() {
		return
	}
	started := time.Now()
	u.congestedSince = started
	timeout := u.queueFullTimeout
	if timeout <= 0 {
		timeout = defaultQueueFullTimeout
	}
	//到时间后执行某段代码
	u.congestionTimer = time.AfterFunc(timeout, func() {
		u.sendMu.Lock()
		// 已恢复的旧计时器不能关闭当前连接。
		if u.closed || u.congestedSince != started {
			u.sendMu.Unlock()
			return
		}
		u.closeSendLocked()
		u.sendMu.Unlock()
		if u.conn != nil {
			_ = u.conn.Close()
		}
	})
}

// 更新活跃时间
func (u *User) UpdateActiveTime() {
	u.activeTime.Store(time.Now().Unix())
}

func (u *User) LastActiveTime() int64 {
	return u.activeTime.Load()
}

// 给当前user对应的客户端发送消息
func (u *User) SendMsg(msg string) error {
	u.sendMu.Lock()
	defer u.sendMu.Unlock()
	if u.closed {
		return fmt.Errorf("user %v: %w", u.ID, ErrConnectionClosed)
	}

	select {
	case u.C <- msg:
		return nil

	default:
		u.startCongestionTimerLocked()
		return fmt.Errorf("user %v: %w", u.ID, ErrSendQueueFull)
	}
}

// 监听某一个用户的信息
// 持续监听 User.C，并将消息发送给客户端。
// 写失败、持续拥塞或外部关闭时退出，通知 Server 清理在线状态。
func (u *User) ListenMessage(disconnect chan<- *User) {
	u.sendMu.Lock()
	done := u.doneLocked()
	u.sendMu.Unlock()
	defer func() {
		u.Close()
		select {
		case disconnect <- u:
		default:
			// 连接已关闭，读循环也会通过 EOF 触发 Offline。
		}
	}()
	for {
		select {
		case <-done:
			return
		case msg, ok := <-u.C:
			if !ok {
				return
			}
			select {
			case <-done:
				return
			default:
			}
			if err := u.conn.Write([]byte(msg + "\n")); err != nil {
				return
			}
			u.sendMu.Lock()
			if !u.closed && len(u.C) < cap(u.C) {
				u.congestedSince = time.Time{}
				if u.congestionTimer != nil {
					u.congestionTimer.Stop()
					u.congestionTimer = nil
				}
			}
			u.sendMu.Unlock()
		}
	}
}

func (u *User) InRoom(roomID int64) bool {
	_, ok := u.JoinedRooms[roomID]
	return ok
}

func (u *User) AddRoom(roomID int64) {
	u.JoinedRooms[roomID] = struct{}{}
}

func (u *User) RemoveRoom(roomID int64) {
	delete(u.JoinedRooms, roomID)
}

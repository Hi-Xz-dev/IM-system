package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval = 20 * time.Second
	pongTimeout  = 60 * time.Second
)

// startHeartbeat runs before the connection's chat read loop starts.
func startHeartbeat(conn *websocket.Conn, interval, timeout time.Duration) (stop func(), err error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(timeout))
	})

	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		//创建一个定时器，每 interval 秒在 ticker.C 上发一次信号
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// WriteControl 可与发送协程的业务消息写入并行。
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}()
	return func() {
		close(done)
		<-stopped
	}, nil
}

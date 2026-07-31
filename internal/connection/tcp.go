package connection

import (
	"net"
	"time"
)

type TCPConnection struct {
	conn net.Conn
}

func NewTCPConnection(conn net.Conn) *TCPConnection {
	return &TCPConnection{
		conn: conn,
	}
}

func (t *TCPConnection) Write(data []byte) error {
	deadline := time.Now().Add(5 * time.Second)
	if err := t.conn.SetWriteDeadline(deadline); err != nil { //本身失败直接返回 说明底层存在问题 等待下面TCP写入 最多5秒
		return err
	}
	_, err := t.conn.Write(data)
	return err
}

func (t *TCPConnection) Close() error {
	return t.conn.Close()
}

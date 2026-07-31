package connection

import (
	"github.com/gorilla/websocket"
)

type WSConnection struct {
	conn *websocket.Conn
}

func NewWSConnection(conn *websocket.Conn) *WSConnection {

	return &WSConnection{
		conn: conn,
	}
}

func (w *WSConnection) Write(data []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, data)

}

func (w *WSConnection) Close() error {
	return w.conn.Close()
}

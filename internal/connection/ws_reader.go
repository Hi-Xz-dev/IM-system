package connection

import "github.com/gorilla/websocket"

type WSReader struct {
	conn *websocket.Conn
}

func NewWSReader(conn *websocket.Conn) *WSReader {
	return &WSReader{
		conn: conn,
	}
}

func (w *WSReader) Read() (string, error) {

	_, msg, err := w.conn.ReadMessage()

	if err != nil {
		return "", err
	}
	return string(msg), nil
}

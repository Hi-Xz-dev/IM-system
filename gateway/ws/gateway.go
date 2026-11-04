package ws

import (
	"IM-system/internal/auth"
	"IM-system/internal/connection"
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"IM-system/server"
	"IM-system/user"

	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	server      *server.Server
	authService *auth.Service
}

func NewGateway(s *server.Server, a *auth.Service) *Gateway {
	return &Gateway{
		server:      s,
		authService: a,
	}
}

// 升级协议
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (g *Gateway) Handler(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	userID, nickname, err := g.Auth(conn)
	if err != nil {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("[系统]认证失败"),
		)

		conn.Close()
		return
	}
	wsConn := connection.NewWSConnection(conn)

	usr :=
		user.NewUser(
			wsConn,
			userID,
			nickname,
			c.Request.RemoteAddr,
		)

	g.server.Online(usr)

	go usr.ListenMessage(g.server.Disconnect)

	usr.SendMsg("[系统] 认证成功")

	reader := connection.NewWSReader(conn)
	//结束通知
	done := make(chan struct{})

	go func() {
		g.server.ServerReader(reader, usr)
		close(done)
	}()

	<-done

	g.server.Offline(usr)

}

func (g *Gateway) Auth(conn *websocket.Conn) (int64, string, error) {
	_, msg, err := conn.ReadMessage()

	if err != nil {
		logger.Log.Warn("authentication failed", "protocol", "websocket", "reason", "read_failed")
		return 0, "", err
	}
	raw := string(msg)
	cmd := protocol.Parse(raw)
	if cmd.Type != domain.CmdAuth {
		logger.Log.Warn("authentication failed", "protocol", "websocket", "reason", "auth_required")
		return 0, "", errors.New("need auth first")
	}

	if len(cmd.Args) != 1 {
		logger.Log.Warn("authentication failed", "protocol", "websocket", "reason", "invalid_format")
		return 0, "", errors.New("invalid auth format")
	}

	token := cmd.Args[0]

	//调用认证服务
	userInfo, err :=
		g.authService.Authenticate(
			context.Background(),
			token,
		)

	if err != nil {
		logger.Log.Warn("authentication failed", "protocol", "websocket", "reason", "authentication_rejected")
		return 0, "", err
	}

	logger.Log.Info("authentication succeeded", "protocol", "websocket", "user_id", userInfo.ID)
	return userInfo.ID, userInfo.Nickname, nil
}

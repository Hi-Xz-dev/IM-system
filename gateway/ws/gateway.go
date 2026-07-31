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
		return 0, "", err
	}
	raw := string(msg)
	logger.Log.Info(
		"auth request",
		"raw", raw,
	)
	cmd := protocol.Parse(raw)
	logger.Log.Info(
		"parsed command",
		"type", cmd.Type,
		"args", cmd.Args,
	)
	if cmd.Type != domain.CmdAuth {
		return 0, "", errors.New("need auth first")
	}

	if len(cmd.Args) != 1 {
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
		logger.Log.Error(
			"authenticate failed",
			"error",
			err,
		)
		return 0, "", err
	}

	return userInfo.ID, userInfo.Nickname, nil
}

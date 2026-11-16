package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"IM-system/gateway/ws"
	"IM-system/internal/auth"
	"IM-system/internal/config"
	"IM-system/internal/database"
	"IM-system/internal/domain"
	"IM-system/internal/httpserver"
	"IM-system/internal/logger"
	"IM-system/internal/message"
	"IM-system/internal/middleware"
	"IM-system/internal/mq"
	"IM-system/internal/repository"
	"IM-system/server"

	"github.com/gin-gonic/gin"
)

func main() {
	logger.Init()
	if err := run(); err != nil {
		logger.Log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	if cfg.MQ.URL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}

	//连接mysql
	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}
	defer db.Close()
	//创建并初始化一个 Server 对象。

	userRepo := repository.NewUserRepository(db)

	roomRepo := repository.NewRoomRepository(db)

	messageRepo := repository.NewMessageRepository(db)

	messageService := message.NewMessageService(messageRepo)

	// 发布与消费都在本服务中，使用各自独立的 RabbitMQ 连接。
	publisher, err := mq.NewRabbitMQ(cfg.MQ.URL, cfg.MQ.QueueName)
	if err != nil {
		return fmt.Errorf("create message publisher: %w", err)
	}
	defer publisher.Close()
	consumer, err := mq.NewConsumer(cfg.MQ.URL, cfg.MQ.QueueName)
	if err != nil {
		return fmt.Errorf("create message consumer: %w", err)
	}
	defer consumer.Close()

	jwtService := auth.NewJWTService(cfg.JWT.Secret)

	authService := auth.NewService(userRepo, jwtService)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	s := server.NewServer(
		cfg.TCP.Host,
		cfg.TCP.Port,
		authService,
		roomRepo,
		messageService,
		publisher,
	)
	// 开放 TCP/HTTP 入口前加载已有房间；加载失败则停止启动。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = s.LoadRooms(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("load rooms: %w", err)
	}

	consumeCtx, stopConsumer := context.WithCancel(context.Background())
	consumerErr := make(chan error, 1)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		consumerErr <- consumer.Run(consumeCtx, func(ctx context.Context, msg domain.Message) error {
			dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return messageService.Persist(dbCtx, msg)
		})
	}()
	defer func() {
		stopConsumer()
		_ = consumer.Close()
		<-consumerDone
	}()
	logger.Log.Info(
		"tcp server starting",
		"host", cfg.TCP.Host,
		"port", cfg.TCP.Port,
	)
	// TCP服务
	go s.Start()

	// gin服务
	//r := gin.Default()
	r := gin.New()

	r.Use(httpserver.Recovery())

	r.Use(httpserver.RequestLogger())

	gateway := ws.NewGateway(s, authService)

	httpserver.RegisterRoutes(
		r,
		s,
		authService,
		gateway,
		authMiddleware,
		roomRepo,
		messageService,
	)
	//静态资源
	r.Static("/web", "./web") //浏览器访问：/web/xxxx去项目中的：./web/xxxx找文件
	HttpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	logger.Log.Info(
		"http server starting",
		"addr", HttpAddr,
	)
	httpServer := &http.Server{Addr: HttpAddr, Handler: r}
	httpErr := make(chan error, 1)
	go func() { httpErr <- httpServer.ListenAndServe() }()
	defer func() {
		logger.Log.Info("server shutting down")
		s.Shutdown()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			_ = httpServer.Close()
		}
	}()

	quit, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-quit.Done():
		return nil
	case err := <-consumerErr:
		// 基础版本不自动重试；退出服务，避免消费者已停而业务继续无声积压。
		return fmt.Errorf("message consumer stopped: %w", err)
	case err := <-httpErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server stopped: %w", err)
	}
}

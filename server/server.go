package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/message"
	"IM-system/internal/model"

	"IM-system/room"
	"IM-system/user"

	"context"
	"io"
	"net"
	"sync"
)

type RoomRepository interface {
	Create(context.Context, *model.Room) error
	FindAll(context.Context) ([]model.Room, error)
}

type MessagePublisher interface {
	Publish(context.Context, domain.Message) error
}

type UserAccountService interface {
	Authenticate(context.Context, string) (*model.User, error)
	UpdateNickname(context.Context, int64, string) (string, error)
}

type Server struct {
	//服务器监听地址
	IP   string
	Port int
	//在线用户列表
	OnlineUsers map[int64][]*user.User
	//保护一组有关联的共享状态
	mapLock sync.RWMutex
	//消息广播的channel
	Message chan string
	//增加房间列表
	Rooms map[int64]*room.Room
	//断开用户链接
	Disconnect chan *user.User
	//服务端的 TCP 监听器
	listener net.Listener
	// 生命周期锁协调关闭状态、监听器和连接登记，不跨越网络 IO。
	lifecycleMu    sync.Mutex
	isShutdown     bool
	done           chan struct{}
	workerWG       sync.WaitGroup
	workersStarted bool
	connections    map[io.Closer]struct{}
	//等待任务完成的计数器
	connectionWG sync.WaitGroup
	//
	authService UserAccountService
	renameMu    sync.Mutex
	//房间信息持久
	roomRepo RoomRepository
	//个人资料持久
	Profiles map[int64]*UserProfile
	//信息持久
	messageService *message.MessageService
	// 消息先由 MQ 确认接收，再异步落库。
	messagePublisher MessagePublisher
}

// 创建一个server接口
func NewServer(
	ip string,
	port int,
	authService UserAccountService,
	roomRepo RoomRepository,
	messageService *message.MessageService,
	messagePublisher MessagePublisher,
) *Server {
	return &Server{
		IP:               ip,
		Port:             port,
		OnlineUsers:      make(map[int64][]*user.User),
		Message:          make(chan string, 100),
		Rooms:            make(map[int64]*room.Room),
		Disconnect:       make(chan *user.User, 100),
		done:             make(chan struct{}),
		connections:      make(map[io.Closer]struct{}),
		authService:      authService,
		roomRepo:         roomRepo,
		Profiles:         make(map[int64]*UserProfile),
		messageService:   messageService,
		messagePublisher: messagePublisher,
	}
}

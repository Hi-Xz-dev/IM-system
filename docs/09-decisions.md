IM-system/
├── cmd/
│   ├── server/
│   │   └── main.go          # 只负责启动
│   └── client/
│       └── main.go
│
├── internal/
│   ├── domain/              # 核心对象
│   │   ├── user.go
│   │   ├── connection.go
│   │   ├── session.go
│   │   ├── message.go
│   │   ├── room.go
│   │   └── command.go
│   │
│   ├── protocol/            # 协议解析
│   │   ├── parser.go
│   │   └── command_type.go
│   │
│   ├── gateway/             # TCP 连接生命周期
│   │   ├── gateway.go
│   │   ├── connection.go
│   │   ├── reader.go
│   │   └── writer.go
│   │
│   ├── hub/                 # 在线连接和消息投递
│   │   ├── hub.go
│   │   ├── registry.go
│   │   └── dispatcher.go
│   │
│   ├── logic/               # 业务逻辑
│   │   ├── service.go
│   │   ├── user_service.go
│   │   ├── room_service.go
│   │   └── chat_service.go
│   │
│   └── httpapi/             # Gin 管理接口
│       ├── router.go
│       ├── handler.go
│       └── dto.go
│
├── docs/
├── web/
├── go.mod
└── README.md
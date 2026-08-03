# Go IM System

> 基于 Go 的实时通信系统，支持 TCP / WebSocket 双协议，从单机 Server 向分布式 IM 演进。

---

## Features

- **TCP / WebSocket Dual Protocol** — 统一 `Connection` 接口，共用 `ServerReader` 和业务逻辑
- **JWT Authentication** — 注册登录 + HTTP 中间件 + WebSocket/TCP 握手认证
- **Multi-Session Management** — 同账号多端在线，房间操作同步所有 session
- **Room Messaging** — 创建/加入/退出房间，房间群聊，进出通知
- **Private Messaging** — 基于用户 ID 的点对点私聊
- **JSON Message Protocol** — `domain.Message` 统一全部消息输出，前端一套解析
- **Concurrent-Safe Message Pipeline** — `SendMsg → User.C → ListenMessage → conn.Write` 单 goroutine 写路径
- **MySQL Persistence** — 用户注册登录，Repository 数据访问层

---

## Quick Start

```bash
# 1. 启动 MySQL
docker run -d --name im-mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=im_system \
  mysql:8

# 2. 启动服务端
go run cmd/server/main.go

# 3. 浏览器打开（建议两个 Tab 用不同账号测试）
open http://localhost:8081/web/
```

默认监听 `127.0.0.1:8080`（TCP）+ `127.0.0.1:8081`（HTTP + WebSocket）。

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                     Client                          │
│          TCP / WebSocket / HTTP REST                │
└────────────────────┬────────────────────────────────┘
                     │
          ┌──────────┴───────────┐
          │   protocol.Parse()   │  ← 文本协议 → Command
          │   DoMessage()        │  ← 命令分发
          └──────────┬───────────┘
                     │
    ┌────────────────┼────────────────┐
    │                │                │
    ▼                ▼                ▼
BroadcastSystem   SendPrivateMsg   SendRoomMsg
  (全服 JSON)      (私聊 JSON)     (房间 JSON)
    │                │                │
    └────────────────┼────────────────┘
                     │
              ┌──────┴──────┐
              │   SendMsg   │  ← 非阻塞，select/default
              │   User.C    │  ← 缓冲 Channel (100)
              │ ListenMessage│  ← 单 goroutine 写
              └──────┬──────┘
                     │
              ┌──────┴──────┐
              │  TCP / WS   │
              └─────────────┘
```

---

## Project Structure

```
IM-system/
├── cmd/server/main.go              # 入口
├── gateway/ws/gateway.go           # WebSocket Gateway
├── internal/
│   ├── auth/                       # JWT + bcrypt + 注册登录
│   ├── config/                     # 配置管理
│   ├── connection/                 # TCP/WS 连接抽象 (Connection + Reader)
│   ├── database/                   # MySQL 连接
│   ├── domain/                     # Command + Message 类型定义
│   ├── httpserver/                 # Gin 路由 + Handler + Middleware + DTO
│   ├── logger/                     # slog 结构化日志
│   ├── middleware/                  # JWT 鉴权中间件
│   ├── model/                      # DB 模型
│   ├── protocol/                   # 协议解析 + JSON 编码
│   └── repository/                 # 数据访问层
├── server/                         # TCP 核心 + 业务逻辑
│   ├── server.go                   # Server struct
│   ├── lifecycle.go                # Start / Shutdown
│   ├── handler.go                  # TCP 连接入口 (认证 → Online)
│   ├── read_loop.go                # ServerReader 统一读循环
│   ├── auth.go                     # TCP 认证握手
│   ├── broadcaster.go              # 全服广播 + ListenMessager
│   ├── message.go                  # SendSystem / SendPrivate / SendRoom
│   ├── command.go                  # DoMessage + 协议 handler
│   ├── room_manager.go             # 房间操作 (TCP + HTTP 共用内核)
│   ├── user_service.go             # 用户操作 (Online/Offline/私聊/改名)
│   └── dto.go                      # RoomInfo / OnlineUser
├── user/user.go                    # User 模型 (多房间, Channel, 连接抽象)
├── room/room.go                    # Room 模型
└── web/index.html                  # Vue 3 聊天控制台
```

---

## Protocol

### Message Format

全部消息统一为 JSON：

```json
{"type":"system","from":0,"from_nickname":"系统","content":"认证成功","time":"..."}
{"type":"system","from":1,"from_nickname":"Tom","content":"hello","time":"..."}
{"type":"private","from":1,"from_nickname":"Tom","to":2,"content":"hi","time":"..."}
{"type":"room","from":1,"from_nickname":"Tom","room_id":"golang","content":"hi","time":"..."}
```

### Authentication

TCP / WebSocket 连接后第一条消息必须发送 `auth|<JWT Token>`，认证成功后进入正常通信。

### Commands

| 命令 | 格式 | 说明 |
|---|---|---|
| 公聊 | 任意文本 | 广播全体在线用户 |
| 私聊 | `to\|用户ID\|内容` | 发送指定用户 |
| 改名 | `rename\|新名字` | 修改昵称并全服广播 |
| 在线列表 | `who` | 查询在线用户 |
| 房间列表 | `rooms` | 所有房间及人数 |
| 创建房间 | `create\|房间名` | 创建新房间 |
| 加入房间 | `join\|房间名` | 加入已有房间 |
| 退出房间 | `leave\|房间名` | 退出房间 |
| 房间群聊 | `room\|房间名\|内容` | 房间内广播 |
| 房间成员 | `members\|房间名` | 查看成员 |
| 当前位置 | `where` | 已加入的房间 |

---

## HTTP API

统一响应 `{"code":0,"msg":"ok","data":...}`。

### Public

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/register` | 注册 |
| POST | `/login` | 登录 → JWT Token |
| GET | `/ping` | 健康检查 |
| GET | `/ws` | WebSocket 升级 |

### Authenticated (`/api`, Bearer Token)

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/online-users` | 在线用户 (ID + 昵称) |
| GET | `/api/rooms` | 房间列表 |
| GET | `/api/rooms/:room/members` | 房间成员 |
| GET | `/api/users/:user/rooms` | 用户已加入的房间 |
| POST | `/api/rooms` | 创建房间 |
| POST | `/api/rooms/:room/members/:user` | 加入房间 |
| DELETE | `/api/rooms/:room/members/:user` | 退出房间 |
| PUT | `/api/user/:user` | 改名 |

---

## Concurrency Model

```
Lock → Snapshot → Unlock → IO

一把 sync.RWMutex 保护 OnlineUsers + Rooms，所有锁不跨越网络 IO
```

| Goroutine | 职责 |
|---|---|
| Accept 循环 | 监听 TCP，启动 Handler |
| `ListenMessager` | 消费 Message chan，全服广播 |
| `CleanOnlineUser` | 定时扫描超时用户 |
| `ListenDisconnect` | 消费 Disconnect chan |
| `ListenMessage` | 消费 User.C，写 TCP/WS（唯一写者） |
| `ServerReader` | 读 TCP/WS，协议分发 |

---

## Roadmap

### Phase 1 — Stability ✅
Handler 生命周期、TCP 粘包/半包、非阻塞广播、协议解析、Offline 幂等、优雅退出、Race Detector、单元测试、Benchmark

### Phase 2 — Engineering ✅
Config、slog、Gin API 分层、JSON 消息协议、JWT 认证、bcrypt、MySQL、Repository、WebSocket Gateway、Connection 接口抽象、多 session 同步、Vue 3 控制台

### Phase 3 — Performance
pprof、压力测试、性能优化

### Phase 4 — Distributed
Redis 用户路由、多 Gateway、跨节点消息、离线消息、消息同步、Docker Compose

---

## License

MIT

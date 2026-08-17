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
- **MySQL Persistence** — 用户 + 房间持久化，`RoomRepository` / `MessageRepository` 数据访问层

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
│   ├── model/                      # DB 模型 (User / Room)
│   ├── protocol/                   # 协议解析 + JSON 编码
│   └── repository/                 # 用户 / 房间 / 消息数据访问层
├── server/                         # TCP 核心 + 业务逻辑
│   ├── server.go                   # Server struct (Rooms/Profiles/roomRepo)
│   ├── lifecycle.go                # Start / Shutdown
│   ├── handler.go                  # TCP 连接入口 (认证 → Online)
│   ├── read_loop.go                # ServerReader 统一读循环
│   ├── auth.go                     # TCP 认证握手
│   ├── broadcaster.go              # 全服广播 + ListenMessager
│   ├── message.go                  # SendSystem / SendPrivate / SendRoom
│   ├── command.go                  # DoMessage + 协议 handler
│   ├── room_manager.go             # 房间操作 (roomID 索引, TCP + HTTP 共用内核)
│   ├── user_service.go             # 用户操作 (Online/Offline/私聊/改名)
│   ├── user_profile.go             # UserProfile (昵称权威来源)
│   └── dto.go                      # RoomInfo / OnlineUser
├── user/user.go                    # User 模型 (多房间, Channel, 连接抽象)
├── room/room.go                    # Room 模型 (ID + Name + Users)
├── migrations/                     # SQL 迁移 (users / rooms)
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
| 创建房间 | `create\|房间名` | 创建新房间（内部写 DB 拿 roomID） |
| 加入房间 | `join\|房间ID` | 加入已有房间 |
| 退出房间 | `leave\|房间ID` | 退出房间 |
| 房间群聊 | `room\|房间ID\|内容` | 房间内广播 |
| 房间成员 | `members\|房间ID` | 查看成员 |
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
| GET | `/api/rooms/:roomID/members` | 房间成员 |
| GET | `/api/users/:user/rooms` | 用户已加入的房间 |
| POST | `/api/rooms` | 创建房间 `{"name":"..."}` |
| POST | `/api/rooms/:roomID/members` | 加入房间 |
| DELETE | `/api/rooms/:roomID/members` | 退出房间 |
| PUT | `/api/user/:user` | 改名 |

---

## Tests

```bash
go test ./...         # 全部测试
go test -race ./...   # + 竞态检测
go test -bench=. -benchmem ./internal/protocol/  # 性能基准
```

| 文件 | 测试函数 | 覆盖内容 |
|---|---|---|
| `internal/auth/service_test.go` | `TestRegister` | 注册 + bcrypt 哈希验证 |
| `internal/auth/service_test.go` | `TestRegisterDuplicate` | 重复注册拒绝 |
| `internal/auth/service_test.go` | `TestLogin` | 登录成功 → JWT Token |
| `internal/auth/service_test.go` | `TestLoginWrongPassword` | 错误密码拒绝 |
| `internal/auth/service_test.go` | `TestAuthenticate` | Token 解析 → 查 DB |
| `internal/protocol/parser_test.go` | `TestParse` | 12 个命令 + 空输入 |
| `internal/protocol/parser_test.go` | `BenchmarkParse` | 26.78 ns/op, 48 B/op, 1 allocs/op |

## Benchmark Results (Apple M4)

| Benchmark | 指标 | 说明 |
|---|---|---|
| `LeaveRoomUnsafe` | 19 ns/op, 0 allocs/op | 状态删除极低成本，无清理瓶颈 |
| `JoinRoomUnsafe` | 376 ns/op, 4 allocs/op | 状态更新 + 成员关系维护开销较小 |
| `BroadcastMessage` (1000 users) | 58.6 µs/op | 广播路径高两个数量级，瓶颈在消息投递 |

**Future Optimization**
- Profile message encoding cost
- Optimize fan-out delivery path
- Evaluate async message dispatch
| `internal/protocol/message_test.go` | `TestEncodeDecodeMessage` | JSON 消息编解码 |
| `server/server_test.go` | `TestOnlineOffline` | 用户上下线 |
| `server/server_test.go` | `TestOfflineDoubleCall` | 下线幂等 (IsClosed) |
| `server/room_manager_test.go` | `TestRoomJoinLeave` | 房间创建/加入/退出/自动销毁 |
| `server/room_manager_test.go` | `TestRoomChat` | 群聊消息投递 |
| `server/room_manager_test.go` | `TestCreateDuplicateRoom` | 重复创建拒绝 |
| `server/room_manager_test.go` | `TestJoinNonExistentRoom` | 加入不存在房间拒绝 |
| `server/room_manager_test.go` | `TestLeaveNonJoinedRoom` | 退出未加入房间拒绝 |
| `server/room_manager_test.go` | `TestConcurrentRoomJoin` | 10 goroutine 并发加入 |
| `server/room_manager_test.go` | `TestConcurrentOffline` | 10 goroutine 并发下线 |
| `server/user_service_test.go` | `TestRenameSync` | 改名 + 房间成员表同步 |
| `server/message_test.go` | `TestPrivateChat` | 私聊消息投递 |
| `server/disconnect_test.go` | `TestListenDisconnect` | 断连通知消费 |
| `user/user_test.go` | `TestUserSendMessage` | SendMsg → ListenMessage 管道 |
| `user/user_test.go` | `TestUserSendMessageQueueFull` | Channel 满返回 error |
| `user/user_test.go` | `TestListenMessage` | 写失败 → Disconnect 通知 |

---

## Data Model

| 结构 | 索引 | 说明 |
|---|---|---|
| `Rooms` | `map[int64]*room.Room` | roomID 索引（DB 自增主键），房间名是业务属性 |
| `OnlineUsers` | `map[int64][]*user.User` | 按 userID 聚合，多端 session 列表 |
| `Profiles` | `map[int64]*UserProfile` | 昵称权威来源，session 不存昵称 |

**昵称迁移**：`Nickname` 从 session 迁到 `Profile`，`Rename` 只需更新 Profile 一处，不再遍历同步所有 session。消息展示时统一 `GetNickname(userID)` 从 Profile 取。

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

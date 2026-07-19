# Go IM System

> 基于 Go 的即时通讯系统，从单机 TCP Server 逐步演进为分布式 IM。

---

## 功能特性

- TCP 长连接 IM 服务（公聊、私聊、房间群聊）
- 多房间支持（用户可同时加入多个房间）
- JWT 认证（注册 / 登录 / Token 校验）
- bcrypt 密码加密
- MySQL 持久化（用户表）
- HTTP REST API（Gin）
- Web 管理控制台（登录注册 + 房间管理）
- 自定义文本协议 + 结构化解析
- 优雅退出（Graceful Shutdown）
- 结构化日志（slog）
- 背压保护（缓冲 Channel + `select/default`）
- TCP 粘包/半包处理（`bufio.Scanner`）
- 统一 TCP 写路径
- Table-driven 单元测试 + Race Detector + Benchmark

---

## 快速启动

```bash
# 1. 启动 MySQL（Docker）
docker run -d --name im-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=im_system mysql:8

# 2. 启动服务端
go run cmd/server/main.go

# 3. 浏览器打开
open http://localhost:8081/web/

# 4. TCP 客户端（可选）
go run cmd/client/main.go
```

默认监听 `127.0.0.1:8080`（TCP）+ `127.0.0.1:8081`（HTTP）。

---

## 项目结构

```
IM-system/
├── cmd/
│   ├── server/main.go              # 入口：组装依赖 + 启动 TCP/HTTP + 信号处理
│   └── client/main.go              # 终端交互客户端
├── internal/
│   ├── auth/
│   │   ├── service.go              # 认证业务（注册 / 登录 / Token 验证）
│   │   ├── jwt.go                  # JWT 生成与解析
│   │   └── password.go             # bcrypt 密码哈希
│   ├── config/config.go            # 配置（TCP / HTTP / MySQL / JWT）
│   ├── database/mysql.go           # MySQL 连接
│   ├── domain/command.go           # 命令类型常量 + Command 结构体
│   ├── httpserver/
│   │   ├── handler.go              # HTTP 处理器（房间 / 用户管理）
│   │   ├── handler_auth.go         # HTTP 处理器（注册 / 登录）
│   │   ├── router.go               # 路由注册
│   │   ├── middleware.go            # 请求日志 + Panic 恢复
│   │   ├── response.go             # 统一 JSON 响应（OK / Fail）
│   │   ├── dto.go                  # 请求 DTO（LoginRequest / RegisterRequest）
│   │   └── params.go               # URL 参数校验
│   ├── logger/logger.go            # 结构化日志（slog）
│   ├── model/user.go               # 数据库 User 模型
│   ├── protocol/
│   │   ├── parser.go               # 文本协议 → Command 解析
│   │   └── parser_test.go          # 12 个命令 + Benchmark
│   ├── repository/
│   │   └── user_repository.go      # 用户数据访问层
│   └── service/
│       └── room_service.go         # Room HTTP 外观层
├── server/                         # TCP 核心服务
│   ├── server.go                   # Server 结构体 + NewServer
│   ├── lifecycle.go                # Start() + Shutdown()
│   ├── handler.go                  # 连接入口（先认证再上线）
│   ├── read_loop.go                # TCP 读循环
│   ├── broadcaster.go              # 消息广播引擎
│   ├── cleaner.go                  # 心跳超时扫描
│   ├── disconnect.go               # 断连通知消费
│   ├── command.go                  # 协议分发（DoMessage + handler）
│   ├── auth.go                     # TCP 认证握手（读第一条消息 → 验证 JWT）
│   ├── room_service.go             # 房间操作（TCP + HTTP 共用内核）
│   ├── user_service.go             # 用户操作（TCP + HTTP 共用内核）
│   ├── dto.go                      # RoomInfo DTO
│   ├── server_test.go              # 测试：上下线、幂等
│   ├── user_service_test.go        # 测试：改名同步
│   └── room_service_test.go        # 测试：房间生命周期
├── user/user.go                    # TCP User 模型（多房间、缓冲 Channel、写超时）
├── room/room.go                    # Room 模型
├── web/index.html                  # Web 管理控制台
├── go.mod
└── README.md
```

---

## 通信协议

TCP 长连接，每条消息以 `\n` 结尾。连接建立后第一条消息必须是认证令牌：

```
auth|<JWT Token>
```

认证成功后进入正常通信：

| 命令 | 格式 | 说明 |
|---|---|---|
| 公聊 | 任意文本 | 广播全体在线用户 |
| 私聊 | `to\|用户名\|内容` | 发送指定用户 |
| 改名 | `rename\|新名字` | 修改昵称（同步所有房间成员表） |
| 在线列表 | `who` | 查询在线用户 |
| 房间列表 | `rooms` | 查看所有房间及人数 |
| 创建房间 | `create\|房间名` | 创建新房间 |
| 加入房间 | `join\|房间名` | 加入已有房间 |
| 退出房间 | `leave\|房间名` | 退出指定房间 |
| 房间群聊 | `room\|房间名\|内容` | 发送到房间全体成员 |
| 房间成员 | `members\|房间名` | 查看房间成员列表 |
| 当前位置 | `where` | 查看已加入的房间 |
| 帮助 | `help` | 打印命令列表 |
| 退出 | `quit` | 断开连接 |

---

## HTTP API

Gin 运行在 `:8081`，统一响应格式：`{"code":0,"msg":"ok","data":...}`。

### 认证

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/register` | 注册 `{"username":"...","password":"...","nickname":"..."}` |
| POST | `/login` | 登录 `{"username":"...","password":"..."}` → 返回 JWT Token |

### 管理

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/ping` | 健康检查 |
| GET | `/online-users` | 查询在线用户名列表 |
| GET | `/rooms` | 查询所有房间及人数 |
| GET | `/rooms/:room/members` | 查询房间成员 |
| GET | `/users/:user/rooms` | 查询用户已加入的房间 |
| POST | `/rooms` | 创建房间 `{"name":"...","user":"..."}` |
| POST | `/rooms/:room/members/:user` | 加入房间 |
| DELETE | `/rooms/:room/members/:user` | 退出房间 |
| PUT | `/users/:user` | 用户改名 `{"name":"新名字"}` |

---

## TCP ↔ HTTP 一致性

```
TCP 命令              HTTP 路由
──────────────────────────────────────────
auth|<token>  →  (TCP 握手认证)
register      →  POST /register
login         →  POST /login
who           →  GET  /online-users
rename|xxx    →  PUT  /users/:user
create|room   →  POST /rooms
rooms         →  GET  /rooms
join|room     →  POST /rooms/:room/members/:user
leave|room    →  DELETE /rooms/:room/members/:user
members|room  →  GET  /rooms/:room/members
where         →  GET  /users/:user/rooms
```

---

## 测试

```bash
go test ./...                        # 全部测试
go test -race ./...                  # + 竞态检测
go test -bench=. -benchmem ./internal/protocol/  # 性能基准
```

| 包 | 测试函数 | 覆盖内容 |
|---|---|---|
| `internal/protocol` | `TestParse` | 12 个命令 + 空输入 |
| `internal/protocol` | `BenchmarkParse` | 26.78 ns/op, 48 B/op |
| `server` | `TestOnlineOffline` | 用户上下线 |
| `server` | `TestOfflineDoubleCall` | 下线幂等（IsClosed） |
| `server` | `TestRoomJoinLeave` | 房间创建/加入/退出/自动销毁 |
| `server` | `TestRenameSync` | 改名同步 OnlineUsers + 房间成员表 |

---

## 认证流程

```
Web 端：                       TCP 端：
┌──────────┐                   ┌──────────┐
│ 注册      │                   │ 登录      │
│ POST     │                   │ Web 端    │
│ /register│                   │ 拿到 Token│
└────┬─────┘                   └────┬─────┘
     │                              │
     ▼                              ▼
┌──────────┐                   ┌──────────┐
│ 登录      │                   │ TCP 连接  │
│ POST     │                   │ 第一条消息 │
│ /login   │                   │ auth|Token│
└────┬─────┘                   └────┬─────┘
     │                              │
     ▼                              ▼
┌──────────────┐              ┌──────────────┐
│ 返回 JWT     │              │ 解析 JWT     │
│ Token        │              │ 查 MySQL     │
│ 存入前端     │              │ 创建 User    │
└──────────────┘              │ 上线 ✅      │
                              └──────────────┘
```

---

## 工程亮点

- **JWT 认证链路** — TCP 握手认证 + HTTP 登录注册，共用 `auth.Service`
- **协议解析抽象** — `Parse(raw) → Command{Type, Args, Raw}` 替代字符串匹配
- **统一 TCP 写路径** — 所有消息经 `User.C → ListenMessage → conn.Write` 单一路径
- **多房间支持** — `JoinedRooms map[string]struct{}`，改名和下线同步所有房间
- **TCP + HTTP 共用内核** — `joinRoomUnsafe` / `leaveRoomUnsafe` 被两套外围共用
- **分层架构** — `handler → service → repository`，依赖注入
- **背压保护** — 缓冲 Channel（容量 100）+ `select/default` 非阻塞发送
- **Handler 生命周期** — `done` channel 保证读协程退出后 Handler 才退出，零泄漏
- **Offline 幂等** — `IsClosed` 标记位防止多路径重复下线
- **优雅退出** — `SIGINT/SIGTERM` → 停止 Accept → 逐用户下线 → 释放资源
- **结构化日志** — `slog`，每条 HTTP 请求记录 method/path/status/cost
- **并发安全** — `sync.RWMutex` + Lock-Snapshot-IO 模式，Race Detector 验证通过

---

## 并发模型

**单锁 + 快照模式：**

```
Lock → 复制数据到本地 Slice → Unlock → 用 Slice 做 IO
  ↑                                      ↑
 数据修改（快）                      不阻塞其他请求
```

一把 `sync.RWMutex` 保护 `OnlineUsers` 和 `Rooms`，所有读写锁不跨越 Channel IO。

### Goroutine 拓扑

| Goroutine | 创建者 | 退出条件 | 职责 |
|---|---|---|---|
| Accept 循环 | `main()` | Shutdown | 监听端口，每连接启动 Handler |
| `ListenMessager` | `Start()` | 进程退出 | 消费 Message chan，snapshot 后广播 |
| `CleanOnlineUser` | `Start()` | 进程退出 | 每 10s 扫描超时用户 |
| `ListenDisconnect` | `Start()` | 进程退出 | 消费 Disconnect chan，触发 Offline |
| `ListenMessage` | `Handler()` | 写超时/写失败 | 消费 User.C，写 TCP（5s 超时） |
| 读协程 | `Handler()` | 连接断开/quit/错误 | 读 TCP，更新心跳，协议分发 |

---

## Roadmap

### Phase 1 — 单机 IM 稳定性 ✅

- [x] Handler 生命周期（零泄漏）
- [x] TCP 粘包/半包
- [x] TCP 写统一出口
- [x] 非阻塞 `select + default`
- [x] 协议解析抽象
- [x] Offline 幂等
- [x] 优雅退出
- [x] Race Detector
- [x] 单元测试
- [x] Benchmark

### Phase 2 — 工程化 ✅

- [x] Config（TCP / HTTP / MySQL / JWT）
- [x] 结构化日志（slog）
- [x] Gin API 分层（`internal/httpserver`）
- [x] 统一 JSON 响应 DTO
- [x] Middleware（RequestLogger + Recovery）
- [x] 多房间模型
- [x] TCP ↔ HTTP API 一致性
- [x] JWT 认证（注册 / 登录 / TCP 握手）
- [x] bcrypt 密码加密
- [x] MySQL 持久化（用户表）
- [x] Repository 模式

### Phase 3 — 性能优化

- pprof（CPU / 内存 / goroutine）
- 压力测试（最大连接数、msg/s、延迟）
- 优化前后数据对比

### Phase 4 — 分布式 IM

- Redis 用户路由
- 多 Gateway 实例
- 跨节点私聊 / 群聊
- 离线消息 / 消息同步
- Docker Compose / Kubernetes

---

## 设计原则

- 先保证正确，再考虑性能
- 先保证稳定，再增加功能
- 每次只解决一个问题
- 每个模块只负责一件事
- 所有重构保持系统可运行
- 所有新功能建立在稳定架构之上

---

## License

MIT

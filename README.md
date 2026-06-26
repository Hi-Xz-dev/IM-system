# IM-system v1.0

基于 Go 的实时即时通讯系统，TCP 长连接 + 自定义文本协议。

## 快速启动

```bash
# 终端1 — 启动服务端
go run cmd/server/main.go

# 终端2 — 启动客户端 A
go run cmd/client/main.go

# 终端3 — 启动客户端 B
go run cmd/client/main.go
```

默认监听 `127.0.0.1:8080`，可通过参数修改：

```bash
go run cmd/server/main.go -ip 0.0.0.0 -port 8888
go run cmd/client/main.go -ip 127.0.0.1 -port 8888
```

## 目录结构

```
IM-system/
├── cmd/
│   ├── server/main.go     # 服务端入口
│   └── client/main.go     # 客户端入口（终端交互）
├── server/
│   ├── server.go          # 生命周期：启动、连接、广播
│   ├── user.go            # 用户操作：上线、下线、改名、私聊
│   ├── room.go            # 房间操作：创建、加入、退出、群聊
│   └── command.go         # 协议层：DoMessage 分发 + handler
├── user/
│   └── user.go            # User 模型
├── room/
│   └── room.go            # Room 模型
└── go.mod
```

## 通信协议

TCP 长连接，每条消息以 `\n` 结尾。

| 指令 | 格式 | 说明 |
|---|---|---|
| 公聊 | `任意文本` | 广播全体在线用户 |
| 私聊 | `to\|用户名\|内容` | 发送指定用户 |
| 改名 | `rename\|新名字` | 修改昵称（同步更新房间成员表）|
| 在线列表 | `who` | 查询在线用户 |
| 房间列表 | `rooms` | 所有房间及人数 |
| 创建房间 | `create\|房间名` | 创建新房间（已在其他房间则自动退出）|
| 加入房间 | `join\|房间名` | 加入已有房间 |
| 退出房间 | `leave` | 退出当前房间 |
| 群聊 | `room\|内容` | 发送到当前房间全体成员 |
| 房间成员 | `members` | 查看当前房间成员列表 |
| 当前位置 | `where` | 查看当前所在房间 |
| 帮助 | `help` | 命令速查 |
| 退出 | `quit` | 断开连接 |

## 客户端使用

```
===== 主菜单 =====
1.公聊模式        ← 任意文本广播
2.私聊模式        ← to|user|msg
3.更新用户名      ← rename|name
4.房间功能        ← 进入子菜单
5.帮助            ← 本地打印命令速查
0.退出            ← 发送 quit 断开

===== 房间功能 =====
1.创建房间  2.加入房间  3.离开房间
4.查看房间  5.当前房间  6.房间成员
7.房间聊天  0.返回
```

## 架构设计

```
┌──────────┐    TCP 长连接    ┌──────────────────┐
│ Client A │ ◄──────────────► │     Server        │
│ (终端)   │                  │                   │
└──────────┘                  │  OnlineUsers      │
                              │  map[name]*User   │
┌──────────┐                  │                   │
│ Client B │ ◄──────────────► │  Rooms            │
│ (终端)   │                  │  map[name]*Room   │
└──────────┘                  │                   │
                              │  Message chan     │
┌──────────┐                  │  (全服广播)        │
│ Client C │ ◄──────────────► │                   │
│ (nc)     │                  └──────────────────┘
└──────────┘
```

## 并发模型

### 单锁 + Snapshot 模式

```
Lock → 复制数据到本地 slice → Unlock → 用 slice 做 IO
  ↑                                      ↑
 数据修改（快）                      不阻塞其他请求
```

一把 `sync.RWMutex` 保护 `OnlineUsers` 和 `Rooms`，所有读写锁不跨越 channel IO。

### Goroutine 拓扑

```
main          → Accept 循环，每连接一个 go Handler
Handler      → conn.Read 循环，阻塞读
ListenMessager → 监听 Message chan，snapshot 后逐用户推送
CleanOnlineUser → 每 10s 扫描超时用户（60s 无心跳踢出）
每个 User      → ListenMessage goroutine（消费 C chan → conn.Write）
```

## 功能特性

- **改名联动**：改名时同步更新房间成员表中的 key
- **房间生命周期**：最后一个用户退出后自动删除房间
- **自动退旧房**：创建/加入新房间时自动离开旧房间
- **超时强踢**：60 秒无心跳自动下线并清理
- **退房广播**：进出房间通知当前房间所有成员
- **下线清理**：用户断开时自动退出房间并从在线表移除

## 测试方式

```bash
# nc 裸调协议
nc 127.0.0.1 8080
rename|alice
create|golang
who

# 客户端交互
go run cmd/client/main.go
```

## 技术栈

- Go 1.26
- 标准库：net、sync、bufio
- 无第三方依赖

# Message Flow

## 1. Purpose

本文档描述 IM 系统中消息的完整流转过程。

消息流定义了消息从客户端发送到最终送达客户端所经过的所有模块。

整个系统中的任何消息都必须遵循统一的消息流。

Message Flow 是整个 IM 系统最重要的设计之一。

---

# 2. Message Flow Overview

所有消息遵循统一处理流程：

```text
                Client
                   │
            TCP Connection
                   │
               Gateway
                   │
              Protocol
                   │
               Command
                   │
                Logic
                   │
                  Hub
                   │
             Connection
                   │
                Client
```

每个模块职责：

| Module | Responsibility |
|----------|----------------|
| Gateway | 网络连接 |
| Protocol | 协议解析 |
| Logic | 业务处理 |
| Hub | 消息投递 |
| Connection | TCP 写出 |

任何消息不得绕过以上流程。

---

# 3. Message Lifecycle

一条消息的生命周期：

```text
Receive

↓

Parse

↓

Validate

↓

Business Process

↓

Route

↓

Send

↓

Complete
```

详细说明：

Receive

Gateway 从 TCP 连接读取原始数据。

↓

Parse

Protocol 将字符串解析为 Command。

↓

Validate

Logic 校验：

- 参数是否合法
- 用户是否在线
- 房间是否存在
- 是否有权限

↓

Business Process

执行业务逻辑。

例如：

- Rename
- Private Chat
- Join Room

↓

Route

Hub 根据业务决定消息发送目标。

例如：

- 单播
- 广播
- 房间广播

↓

Send

Connection 将消息写入客户端。

↓

Complete

消息生命周期结束。

---

# 4. Broadcast Flow

客户端发送：

```text
hello
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Hub.Broadcast()

↓

All Connections

↓

Clients
```

说明：

Broadcast 不直接操作 TCP。

Broadcast 只负责生成广播事件。

真正写 Socket 的只有 Connection。

---

# 5. Private Chat Flow

客户端发送：

```text
to|Tom|hello
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Find User

↓

Hub.Send()

↓

Tom Connection

↓

Tom Client
```

Logic：

负责：

- 查询用户
- 权限校验

Hub：

负责：

- 消息投递

---

# 6. Room Chat Flow

客户端发送：

```text
room|hello
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Find Room

↓

Room Members

↓

Hub.Broadcast(Room)

↓

Connections

↓

Clients
```

Room 负责：

成员管理。

Hub 负责：

消息发送。

---

# 7. Join Room Flow

客户端发送：

```text
join|room1
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Check Room

↓

Update Room

↓

Hub.Broadcast(System Message)

↓

Room Members
```

Logic：

修改房间状态。

Hub：

通知所有成员。

---

# 8. Leave Room Flow

客户端发送：

```text
leave
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Remove Member

↓

Hub.Broadcast(System Message)

↓

Room Members
```

如果房间为空：

Logic 自动删除 Room。

---

# 9. Rename Flow

客户端发送：

```text
rename|Tom
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Update User

↓

Hub.Send(System Message)

↓

Client
```

Logic：

检查：

- 重名
- 用户状态

---

# 10. Online Query Flow

客户端发送：

```text
who
```

消息流：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Query Online Users

↓

Hub.Send()

↓

Client
```

查询类请求不会广播。

---

# 11. HTTP Flow

HTTP API 不属于聊天消息流。

HTTP 请求流程：

```text
Browser

↓

Gin

↓

Logic

↓

Hub / Domain

↓

JSON Response
```

HTTP 不直接访问：

- TCP Connection
- net.Conn
- Socket

---

# 12. Future Distributed Flow

未来支持分布式后：

消息流升级为：

```text
Client

↓

Gateway

↓

Protocol

↓

Logic

↓

Hub

↓

Redis Router

↓

Remote Gateway

↓

Connection

↓

Remote Client
```

新增：

Redis：

负责：

用户路由。

Gateway：

负责：

跨节点连接。

Logic：

无需修改。

---

# 13. Design Rules

所有消息必须遵循以下原则：

## Rule 1

Gateway 不处理业务。

---

## Rule 2

Protocol 不处理业务。

---

## Rule 3

Logic 不直接操作 TCP。

---

## Rule 4

Hub 是唯一消息出口。

任何 TCP 写操作最终必须经过 Hub。

---

## Rule 5

Connection 只负责 Socket IO。

Connection 不处理业务。

---

# 14. Future Evolution

未来 Message Flow 将继续扩展：

Current

```text
Client

↓

Gateway

↓

Logic

↓

Hub

↓

Client
```

↓

Engineering

增加：

- Middleware
- Logger

↓

Performance

增加：

- Message Queue
- Metrics

↓

Distributed

增加：

- Redis
- Multiple Gateway
- Service Discovery

↓

Cloud Native

增加：

- Kubernetes
- Load Balancer
- Distributed Message Queue

但整个消息流的核心思想保持不变：

Receive

↓

Parse

↓

Logic

↓

Route

↓

Send
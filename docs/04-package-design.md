# Package Design

## Overview

整个项目采用分层设计。

每个 Package 只负责一类职责。

Package 之间遵循单向依赖原则。

# Dependency Rule

系统依赖关系如下：

```text
cmd
 ├── gateway
 │     └── protocol
 │
 ├── httpapi
 │
 └── logic
       ├── hub
       ├── domain
       └── storage（Future）

任何 Package 不允许反向依赖。

---

# Package Overview

| Package | Responsibility |
|----------|----------------|
| gateway | TCP连接生命周期 |
| protocol | 协议解析 |
| logic | IM业务逻辑 |
| hub | 在线连接与消息投递 |
| httpapi | HTTP接口 |
| storage | 数据存储（Future） |

---

# gateway

## Responsibility

Gateway 负责所有网络连接。

包括：

- TCP Listen
- Accept
- Read
- Write
- Heartbeat
- Disconnect
- Graceful Shutdown

Gateway 只负责连接。

不负责任何业务。

---

## Owns

Gateway 拥有：

- net.Listener
- net.Conn
- Connection 生命周期

---

## Does NOT Own

Gateway 不负责：

- 用户
- 房间
- 广播策略
- 私聊
- 登录

---

# protocol

## Responsibility

负责协议解析。

例如：

```
rename|Tom
```

↓

```
Command{
    Type: Rename,
    Args:["Tom"]
}
```

Protocol 不关心：

- 用户是否存在
- 房间是否存在
- 消息是否成功发送

---

## Owns

- Parser
- Command
- Message Format

---

## Does NOT Own

- Business Logic
- Connection
- User

---

# logic

## Responsibility

负责 IM 的所有业务。

包括：

- Rename
- Who
- Private Chat
- Create Room
- Join Room
- Leave Room
- Room Chat

Logic 不直接操作 TCP。

Logic 不直接操作 net.Conn。

---

## Owns

- User Business
- Room Business
- Permission Check
- Command Dispatch

---

## Does NOT Own

- Socket
- Listener
- HTTP

---

# hub

## Responsibility

负责消息投递。

所有消息最终都经过 Hub。

包括：

- Broadcast
- Send To User
- Send To Room
- Online User Registry

以后：

任何 TCP 写操作。

必须统一经过 Hub。

---

## Owns

- Online Connections
- Message Queue
- Connection Registry

---

## Does NOT Own

- 协议解析
- 房间业务
- 登录业务

---

# httpapi

## Responsibility

提供 HTTP 管理接口。

例如：

- Online Users
- Rooms
- Members

HTTP API 不直接访问连接。

HTTP API 仅调用 Logic。

---

# storage（Future）

当前阶段暂未实现。

未来负责：

- MySQL
- Redis
- Message History
- Offline Message

Storage 不关心业务。

仅负责数据持久化。

---

# Dependency Rule

系统依赖关系如下：

```
Gateway

↓

Protocol

↓

Logic

↓

Hub

↓

Storage
```

所有依赖只能向下。

禁止：

```
Hub

↓

Gateway
```

禁止：

```
Storage

↓

Logic
```

---

# Refactoring Rule

新增代码前必须回答：

1.

这个功能属于哪个 Package？

2.

这个 Package 是否已经承担了其它职责？

3.

是否违反单一职责原则？

如果不能回答以上三个问题，

则不能开始编码。
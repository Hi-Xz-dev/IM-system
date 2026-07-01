# Domain Model

## 1. Purpose

本文档用于定义 IM 系统中的核心领域对象。

领域模型用于指导后续 package 设计、数据库设计、消息流设计和分布式改造。

---

## 2. Core Domain Objects

当前系统包含以下核心对象：

- User
- Connection
- Session
- Message
- Room
- Command

---

## 3. User

User 表示一个用户身份。

User 不等于 Connection。

一个 User 未来可以同时拥有多个 Connection。

### Responsibilities

- 表示用户身份
- 保存用户基础信息
- 维护用户在线状态

### Fields

- ID
- Name
- Status
- CreatedAt

---

## 4. Connection

Connection 表示一个网络连接。

当前阶段是 TCP 连接，未来可以扩展为 WebSocket。

### Responsibilities

- 封装底层网络连接
- 负责读写消息
- 维护心跳状态
- 处理断开事件

### Fields

- ID
- Conn
- RemoteAddr
- State
- LastActiveAt

---

## 5. Session

Session 表示 User 和 Connection 的绑定关系。

登录成功后，系统创建 Session。

### Responsibilities

- 绑定 User 和 Connection
- 记录登录状态
- 管理会话生命周期

### Fields

- ID
- UserID
- ConnectionID
- LoginAt
- State

---

## 6. Message

Message 表示系统中的一条消息。

私聊、群聊、系统通知，本质上都可以抽象为 Message。

### Responsibilities

- 表示消息内容
- 表示消息发送方
- 表示消息接收方
- 表示消息类型

### Fields

- ID
- From
- To
- Type
- Content
- CreatedAt

---

## 7. Room

Room 表示一个群聊空间。

当前阶段用于房间聊天，未来可以演进为群组、频道或讨论组。

### Responsibilities

- 管理房间成员
- 定义消息作用域
- 支持房间广播

### Fields

- ID
- Name
- Members
- OwnerID
- State

---

## 8. Command

Command 表示客户端发送的协议命令。

例如：

```text
join|room1
to|Tom|hello
room|hello
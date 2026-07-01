# Architecture

## 1. Vision

本项目旨在实现一个高性能、高可扩展性的即时通信系统。

项目采用渐进式演进方式：

Standalone IM

↓

Engineering IM

↓

High Performance IM

↓

Distributed IM

最终实现支持多实例部署的 Distributed Instant Messaging System。

---

## 2. Design Principles

整个系统遵循以下原则：

### 2.1 Single Responsibility

每个模块只负责一件事情。

### 2.2 High Cohesion

模块内部高度聚合。

### 2.3 Low Coupling

模块之间通过接口通信。

### 2.4 Message Driven

所有消息流采用统一消息模型。

### 2.5 Testability

所有业务逻辑必须可测试。

### 2.6 Evolvability

所有模块能够独立演进。

---

## 3. Overall Architecture

```text
                 Client
                    │
              TCP Connection
                    │
             Gateway Layer
                    │
          Protocol Dispatcher
                    │
             Business Logic
                    │
               Message Hub
          ┌─────────┴─────────┐
      User Manager      Room Manager
```
---

```text
4. Layer Responsibilities
4.1 Gateway Layer

负责 TCP 连接生命周期：

Accept
Read
Write
Heartbeat
Disconnect
Graceful Shutdown

Gateway 不处理业务逻辑。

4.2 Protocol Layer

负责协议解析：

文本命令解析
参数校验
Command 封装

Protocol 不关心用户是否在线，也不关心房间是否存在。

4.3 Business Logic Layer

负责 IM 核心业务：

用户改名
在线用户查询
私聊
房间创建
加入房间
离开房间
房间群聊

Logic 不直接操作 TCP 连接。

4.4 Message Hub

负责消息投递：

单用户投递
全服广播
房间广播
在线连接管理

所有 TCP 写操作最终都应该通过 Hub 统一出口。

4.5 User Manager

负责用户状态管理：

用户上线
用户下线
用户改名
在线用户查询
4.6 Room Manager

负责房间状态管理：

创建房间
加入房间
离开房间
查询房间成员
房间自动销毁

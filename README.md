# Go IM System

> A high-performance Instant Messaging System built with Go.

---

# Project Overview

Go IM System 是一个基于 Go 开发的即时通信系统。

项目最终目标不是实现一个简单的聊天室，而是逐步演进为一个具备工程化能力、可扩展、支持多实例部署的 Distributed IM System。

整个项目将按照真实后端系统的开发流程持续迭代，而不是一次性完成所有功能。

---

# Design Philosophy

本项目遵循以下原则：

- 先保证正确，再考虑性能
- 先保证稳定，再增加功能
- 每次只解决一个问题
- 每个模块只负责一件事情
- 所有重构必须保持系统可运行
- 所有新增功能必须建立在稳定架构之上

---

# Development Roadmap

## Phase 1：Standalone IM

目标：

构建一个稳定、正确、可测试的单机 IM Server。

重点完成：

- Handler 生命周期管理
- TCP 粘包 / 半包处理
- SendMsg 重构
- 所有 TCP 写统一出口
- 优雅退出（Graceful Shutdown）
- race detector
- Unit Test
- Benchmark

完成本阶段后：

项目可以作为 Go 后端实习项目。

---

## Phase 2：Engineering

目标：

让项目具备真实后端项目结构。

包括：

- Gin API 分层
- DTO
- Middleware
- Config
- Logger
- JWT
- bcrypt
- MySQL
- Repository

数据库主要用于：

- 用户登录
- 历史消息
- 离线消息

---

## Phase 3：Performance

目标：

用数据证明系统性能。

包括：

- pprof
- Benchmark
- 压测
- 最大连接数
- Messages / Second
- CPU Analysis
- Memory Analysis
- Performance Optimization

---

## Phase 4：Distributed IM

目标：

演进为分布式即时通信系统。

包括：

- Redis
- Online User Routing
- Multi Gateway
- Cross-node Private Chat
- Cross-node Room Chat
- Offline Message
- Message Synchronization
- Docker Compose
- Kubernetes（Optional）

完成本阶段后：

项目正式升级为：

Distributed IM System

---

# Why Refactor?

项目第一版已经实现：

- 用户上线/下线
- 公聊
- 私聊
- 房间聊天
- HTTP API

随着功能增加，Server 同时负责：

- TCP 连接管理
- 用户管理
- 房间管理
- 消息广播
- 命令解析
- HTTP API

导致：

- 模块职责越来越多
- 耦合越来越严重
- 可维护性下降
- 后续开发成本越来越高

因此决定停止增加新功能，开始重构整个项目。

重构目标不是修改功能，而是建立能够持续演进的架构。

---

# Refactoring Principles

重构期间遵循以下原则：

- 不新增聊天功能
- 不修改通信协议
- 每次只解决一个问题
- 每次重构后保证项目可运行
- 保持提交粒度尽可能小
- 每完成一个阶段再进入下一阶段

---

# Current Stage

当前项目处于：

Phase 1：Standalone IM

当前工作重点：

- [ ] Handler 生命周期
- [ ] TCP 粘包 / 半包
- [ ] SendMsg 重构
- [ ] TCP 写统一出口
- [ ] Graceful Shutdown
- [ ] race detector
- [ ] Unit Test
- [ ] Benchmark

---

# Long-term Architecture

```
                Client
                   │
            TCP / WebSocket
                   │
              Connection
                   │
              Gateway Layer
                   │
              Protocol Layer
                   │
               Logic Layer
                   │
              Message Hub
                   │
        ┌──────────┴──────────┐
        │                     │
      Storage              Cache
      MySQL               Redis
```

当前阶段仅实现：

```
Client
    │
TCP Connection
    │
Standalone IM Server
```

后续逐步演进为完整的分布式架构。

---

# License

MIT

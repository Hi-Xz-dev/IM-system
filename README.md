# Go IM System

> A high-performance Instant Messaging System built with Go.

---

# Features

- TCP-based Instant Messaging
- Public Chat
- Private Chat
- Room Chat
- HTTP REST API (Gin)
- Graceful Shutdown
- Sticky Packet / Half Packet Handling
- Protocol Parser
- Unified TCP Write Path
- Unit Tests
- Race Detector

---

# Project Overview

Go IM System is an Instant Messaging System built with Go.

The goal of this project is **not** to build a simple chat room, but to gradually evolve into a production-oriented **Distributed IM System**.

Instead of focusing only on features, this project follows an incremental engineering roadmap, continuously improving architecture, maintainability, and scalability.

---

# Design Philosophy

This project follows several engineering principles:

- Correctness before performance
- Stability before features
- Solve one problem at a time
- One module, one responsibility
- Every refactoring keeps the system runnable
- Every new feature is built on a stable architecture

---

# Development Roadmap

## Phase 1 — Standalone IM

Goal:

Build a stable, correct and testable standalone IM server.

- [x] Handler lifecycle management (no goroutine leaks)
- [x] TCP sticky packet / half packet handling (`bufio.Scanner`)
- [x] Unified TCP write path (`User.C -> ListenMessage`)
- [x] SendMsg refactoring (non-blocking `select + default`)
- [x] Protocol.Parse abstraction (`internal/protocol`)
- [x] Command abstraction (`internal/domain`)
- [x] Offline idempotency (`IsClosed`)
- [x] Graceful Shutdown (`SIGINT/SIGTERM`)
- [x] Go Race Detector (`go test -race`)
- [x] Unit Tests
- [x] Benchmark

---

## Phase 2 — Engineering

Goal:

Transform the project into a production-style Go backend.

Including:

- Gin API layering
- DTO
- Middleware
- Config
- Logger
- JWT Authentication
- bcrypt
- MySQL
- Repository Pattern

Database will mainly be used for:

- User authentication
- Message history
- Offline messages

---

## Phase 3 — Performance

Goal:

Measure and optimize system performance with real data.

Including:

- pprof
- Benchmark
- Stress Testing
- Maximum Connections
- Messages / Second
- CPU Profiling
- Memory Profiling
- Performance Optimization

---

## Phase 4 — Distributed IM

Goal:

Evolve into a distributed instant messaging system.

Including:

- Redis
- Online User Routing
- Multi-Gateway
- Cross-node Private Chat
- Cross-node Room Chat
- Offline Messages
- Message Synchronization
- Docker Compose
- Kubernetes (Optional)

After this phase, the project officially becomes:

**Distributed IM System**

---

# Engineering Highlights

- **Protocol Parser Abstraction** — `Parse(raw) -> Command{Type, Args, Raw}` replaces string-based command matching.
- **Unified TCP Write Path** — All outgoing messages go through `User.C -> ListenMessage -> conn.Write`.
- **Graceful Shutdown** — `SIGINT/SIGTERM -> Stop Accept -> Offline Users -> Release Resources`.
- **Handler Lifecycle Management** — `done` channel guarantees Handler exits only after the reader goroutine exits.
- **Sticky Packet Handling** — `bufio.Scanner` reads messages line by line, eliminating TCP sticky/half packet issues.
- **Offline Idempotency** — `IsClosed` prevents duplicate offline operations.
- **Backpressure Protection** — Buffered channels with `select/default` prevent slow consumers from blocking the server.
- **Unit Testing** — Table-driven tests cover protocol parsing and core server logic.
- **Concurrency Safety** — Shared state is protected by `sync.RWMutex` and verified using Go Race Detector.

---

# Testing

```bash
# Run all tests
go test ./...

# Run all tests with race detector
go test -race ./...

# Run protocol tests
go test -v ./internal/protocol
```

## Test Coverage

| Package | Test | Description |
|----------|------|-------------|
| `internal/protocol` | `TestParse` | Protocol parsing (12 commands + empty input) |
| `server` | `TestOnlineOffline` | User online/offline lifecycle |
| `server` | `TestOfflineDoubleCall` | Offline idempotency |
| `server` | `TestRoomJoinLeave` | Room lifecycle |
| `server` | `TestRenameSync` | Rename synchronization |

---

# Why Refactor?

The first version of the project already supported:

- User online/offline
- Public chat
- Private chat
- Room chat
- HTTP API

As more features were added, the Server gradually became responsible for:

- TCP connection management
- User management
- Room management
- Message broadcasting
- Protocol parsing
- HTTP API

This resulted in:

- Increasing responsibilities
- High coupling
- Reduced maintainability
- Higher development cost

Therefore, feature development was intentionally paused in favor of architectural refactoring.

The goal of refactoring is **not** to add more features, but to improve maintainability, testability, and long-term scalability.

---

# Refactoring Principles

During refactoring:

- No new chat features
- No protocol changes
- One problem solved at a time
- Keep the project runnable after every refactoring
- Keep commits small and focused
- Move to the next stage only after finishing the current one

---

# Current Stage

Current status:

**Phase 1 (Completed) → Phase 2 (In Progress)**

Phase 1 has achieved its primary goal:

- Stable architecture
- Correct behavior
- Engineering-oriented refactoring
- Unit testing

Current focus:

- [ ] Gin API Layer
- [ ] DTO
- [ ] Middleware
- [ ] Config
- [ ] Logger
- [ ] JWT Authentication
- [ ] MySQL
- [ ] Repository

---

# Long-term Architecture

```
                Client
                   │
            TCP / WebSocket
                   │
              Gateway Layer
                   │
             Protocol Layer
                   │
           Application Layer
                   │
             Message Hub
                   │
        ┌──────────┴──────────┐
        │                     │
      Storage              Cache
      MySQL               Redis
```

Current implementation:

```
Client
    │
TCP Connection
    │
Standalone IM Server
```

Future versions will gradually evolve into a distributed architecture.

---

# License

MIT

# IM System

基于 Go 的即时通信项目，支持 TCP、WebSocket 和 HTTP API，附带 Vue 3 网页聊天界面。

项目定位为在线聊天室：任意端加入或主动退房，同步同账号所有在线端；新端登录继承其他在线端的房间。单端断线不影响其他端，所有端离线后不保留成员身份，再登录需要重新加入。房间无人时保留，服务器启动时从数据库加载房间资料；暂不设置房主，也不保存永久成员关系。

每个连接使用容量 100 的发送队列，满时拒绝本次投递；持续拥塞 5 秒未恢复则关闭该连接，TCP/WS 网络写入超时均为 5 秒。入队成功不代表客户端已经收到，当前不提供自动重试或断线补拉保证。

## 功能

- 用户注册、登录与 JWT 认证
- 公聊、私聊、房间创建与群聊
- 同账号多端在线、在线用户查询与昵称修改
- MySQL 用户、房间和消息持久化，以及聊天历史查询

## 目录结构

```text
cmd/server/       服务启动入口
gateway/         WebSocket 网关
internal/         认证、配置、连接、协议、HTTP 接口、业务服务和数据库访问
server/           聊天核心、用户会话、房间管理和消息分发
user/             用户会话模型与消息发送队列
room/             房间模型
migrations/       数据库建表脚本
web/              网页聊天界面
docs/            HTTP 接口和通信命令说明
```

## 本地运行

准备 Go 1.26.3 或更高版本（见 `go.mod`）、MySQL 8 和 RabbitMQ（支持 quorum 队列）。以下命令均在项目根目录执行。

1. 创建数据库并按顺序导入表结构（已有表时无需重复执行）：

   ```bash
   mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS im_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
   mysql -u root -p im_system < migrations/001_create_users.sql
   mysql -u root -p im_system < migrations/002_create_rooms.sql
   mysql -u root -p im_system < migrations/003_create_messages.sql
   ```

2. 检查 [internal/config/config.go](internal/config/config.go) 中的 MySQL 配置。默认使用 `root`、空密码连接本地数据库，仅在与你的实际配置不一致时修改；启动 MySQL 服务不等于登录或设置密码，Go 程序会自行连接。JWT 密钥通过环境变量 `JWT_SECRET` 提供，至少 32 字节，未设置或过短时服务会停止启动。使用 `openssl rand -hex 32` 生成随机密钥后，在启动服务的终端设置 `export JWT_SECRET='<替换为生成的随机密钥>'`；将实际密钥安全保存，不提交到 Git，后续启动保持相同值。更换密钥后，旧 Token 将无法验证。

   其余默认配置如下：

   | 配置 | 默认值 |
   | --- | --- |
   | TCP 地址 | `127.0.0.1:8080` |
   | HTTP / WebSocket 地址 | `127.0.0.1:8081` |
   | MySQL 地址 | `127.0.0.1:3306` |
   | 数据库 | `im_system` |
   | 数据库账号 / 密码 | `root` / 空密码 |

3. 启动 RabbitMQ，设置连接地址后启动 Go 服务（本地 Homebrew 安装可使用 `brew services start rabbitmq`）：

   ```bash
   go mod download
   export RABBITMQ_URL='amqp://guest:guest@127.0.0.1:5672/'
   go run ./cmd/server
   ```

4. 打开 [聊天页面](http://localhost:8081/web/)，注册并登录。可用两个标签页登录不同账号测试聊天。页面从 CDN 加载 Vue，需要联网。

## 测试

私聊与房间消息先发布到 RabbitMQ 的 `im.message.persist` 队列，确认接收且未被退回后再投递给在线用户。同一 Go 服务内的消费 goroutine 调用 `Persist` 写入 MySQL，成功后手动 ACK；消息 ID 唯一键和内容核对用于处理重复投递。系统通知与全服广播暂不落库。

只有一个 Go 启动入口，无需另开发布或消费程序。发布确认不代表已经落库或客户端已收到；消费者出错时服务会退出，未确认消息留待后续投递，需要排查后手动重启，当前不提供自动重连、重试和死信处理。单节点 quorum 队列不具备多节点容灾能力。

验证时登录两个账号，发送私聊或房间消息，再在 MySQL 查询 `SELECT id, message_id, content FROM messages ORDER BY id DESC LIMIT 10;`。可停止整个 Go 服务后重启，检查已有消息不重复；这不能代替“落库后、ACK 前中断”的故障实验。

```bash
go test ./...
go test -race ./...
```

可选的真实 MQ 集成测试使用专属临时队列，验证聊天发送、消费、幂等落库和 ACK，结束时删除测试队列与记录：

```bash
RABBITMQ_TEST_URL='amqp://guest:guest@127.0.0.1:5672/' go test ./server -run '^TestChatMQPersistence$' -v
```

认证和消息仓储测试需要本地 MySQL，使用 `root` 空密码连接 `127.0.0.1:3306/im_system`，并需要提前建表；这些测试会写入并清理测试数据。

接口路径和 TCP / WebSocket 命令见 [接口与协议说明](docs/api.md)。当前问题、复现步骤和验收标准见 [问题排查清单](docs/troubleshooting.md)。

## 许可证

[MIT](LICENSE)

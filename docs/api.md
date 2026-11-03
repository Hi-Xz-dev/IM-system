# 接口与协议

## 聊天室边界

- 加入房间以账号为单位，HTTP、TCP 和 WebSocket 命令都会同步同账号所有在线连接。
- 新连接上线时，继承同账号其他在线连接已加入的房间。
- 主动退房使同账号所有端退出；断线只清理当前连接。
- 所有端离线后不保留账号的房间成员身份，再登录需重新加入；房间没有在线成员时删除运行时状态。
- 数据库中的房间记录和历史消息不代表永久成员关系；数据库房间记录的清理和同名重建问题仍见排查清单第 06 项。

## HTTP API

默认地址：`http://127.0.0.1:8081`。响应中的 `code` 为 `0` 表示成功，`-1` 表示失败；`data` 为空时可能省略。

```json
{"code":0,"msg":"ok","data":{}}
```

### 公开接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/register` | 注册，JSON 字段为 `username`、`password`、`nickname` |
| POST | `/login` | 登录，JSON 字段为 `username`、`password`，返回 JWT 等登录信息 |
| GET | `/ping` | 健康检查 |
| GET | `/ws` | WebSocket 连接入口，升级后发送认证命令 |

### 登录后接口

以下接口需要请求头 `Authorization: Bearer <token>`。`:room` 和 `:user` 分别为数字房间 ID 和用户 ID。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/online-users` | 在线用户列表 |
| GET | `/api/rooms` | 房间列表 |
| GET | `/api/users/:user/rooms` | 当前登录用户已加入的房间 |
| GET | `/api/rooms/:room/members` | 房间成员 |
| POST | `/api/rooms` | 创建房间，JSON 为 `{"name":"房间名"}` |
| POST | `/api/rooms/:room/members/:user` | 当前登录用户加入房间 |
| DELETE | `/api/rooms/:room/members/:user` | 当前登录用户退出房间 |
| PUT | `/api/user/:user` | 当前登录用户改名，JSON 为 `{"name":"新昵称"}` |
| GET | `/api/rooms/:room/history` | 房间最近 20 条消息 |
| GET | `/api/private/:user/history` | 与指定用户的最近 20 条私聊消息 |

当前实现中，查询已加入房间、加入/退出房间和改名操作以 JWT 中的用户身份为准，路径中的 `:user` 不决定操作对象；私聊历史的 `:user` 则表示对方用户。

## TCP / WebSocket 命令

TCP 连接地址为 `127.0.0.1:8080`，每条命令以换行结束；WebSocket 地址为 `ws://127.0.0.1:8081/ws`，每条命令作为一条文本消息发送。

连接后的第一条消息必须为 `auth|<token>`，其中 token 由 HTTP 登录接口获取。

| 命令 | 格式 |
| --- | --- |
| 认证 | `auth\|<token>` |
| 公聊 | 直接发送文本 |
| 私聊 | `to\|用户ID\|内容` |
| 修改昵称 | `rename\|新昵称` |
| 在线用户 | `who` |
| 房间列表 | `rooms` |
| 创建房间 | `create\|房间名` |
| 加入房间 | `join\|房间ID` |
| 退出房间 | `leave\|房间ID` |
| 房间群聊 | `room\|房间ID\|内容` |
| 房间成员 | `members\|房间ID` |
| 已加入的房间 | `where` |
| 房间历史 | `history-room\|房间ID` |
| 私聊历史 | `history-private\|用户ID` |
| 帮助 | `help` |

聊天消息使用 JSON，房间 ID 为数字，例如：

```json
{"id":"","type":"room","from":1,"from_nickname":"Tom","room_id":1,"room_name":"Go","content":"hello","time":"2026-10-21T12:00:00+08:00"}
```

认证提示等部分系统输出仍为纯文本，客户端需要兼容处理。

[返回 README](../README.md)

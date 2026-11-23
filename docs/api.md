# 接口与协议

## 聊天室边界

- 加入房间以账号为单位，HTTP、TCP 和 WebSocket 命令都会同步同账号所有在线连接。
- 新连接上线时，继承同账号其他在线连接已加入的房间。
- 主动退房使同账号所有端退出；断线只清理当前连接。
- 所有端离线后不保留账号的房间成员身份，再登录需重新加入；房间没有在线成员时仍保留，启动时从数据库加载房间 ID、名称，在线成员初始化为空。
- 数据库中的房间记录和历史消息不代表永久成员关系；暂不设置房主或解散权限，同名房间已存在时应加入原房间。
- 已登录但未建立聊天连接时也可创建空房；有在线连接时创建后同步加入所有端。

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
| GET | `/api/me/rooms` | 当前登录用户已加入的房间 |
| GET | `/api/rooms/:room/members` | 当前成员可查看名单，按账号返回 `id` 和 `nickname` |
| POST | `/api/rooms` | 创建房间，JSON 为 `{"name":"房间名"}` |
| POST | `/api/rooms/:room/membership` | 当前登录用户加入房间 |
| DELETE | `/api/rooms/:room/membership` | 当前登录用户退出房间 |
| PUT | `/api/me/nickname` | 当前登录用户改名，JSON 为 `{"name":"新昵称"}` |
| GET | `/api/rooms/:room/history` | 当前成员可查看房间最近 20 条消息 |
| GET | `/api/private/:user/history` | 与指定用户的最近 20 条私聊消息 |

查询已加入的房间、入房、退房和改名均使用 JWT 中的用户身份。私聊历史的 `:user` 表示对方用户。原路径 `/api/users/:user/rooms`、`/api/user/:user` 和 `/api/rooms/:room/members/:user` 保留为兼容入口，操作对象仍由 JWT 决定。

登录后可以列出和加入公开房间。HTTP、TCP 和 WebSocket 都要求当前账号已入房才能查看成员和历史，历史默认返回最近 20 条，包含入房前的消息。主动退房或账号所有端离线后，须重新加入才能查看。HTTP 对未入房返回 403，对不存在的房间返回 404。

房间和私聊的 HTTP 历史接口默认返回最新一页，每页最多 20 条，`data` 仍是消息数组。有更早记录时，响应头 `X-Next-Cursor` 会给出下一页游标。下一次请求在同一接口加 `?before=<游标>`，继续从上一页末尾往前查。没有这个响应头就到末页了。无效游标返回 400，每页仍执行原有的访问权限检查。各页读取查询时已落库的记录，不共享数据库快照。

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

成功执行 `history-room` 或 `history-private` 会收到一条 JSON 响应，外层 `type` 为 `history`。房间历史带 `room_id`，私聊历史带 `peer_id`。例如房间历史为空时返回：

```json
{"type":"history","room_id":1,"messages":[]}
```

`messages` 包含最近最多 20 条消息，按从旧到新排列，每条保留原始消息 ID、发送者、类型和时间。TCP 客户端需要能读入包含整页历史的一行 JSON。

历史命令失败时返回 `type: "error"`。客户端可以用 `command` 确定失败的命令，用 `code` 判断原因，并把 `message` 显示给用户。

```json
{"type":"error","command":"history-room","code":"not_room_member","message":"当前未加入房间"}
```

错误码包括 `invalid_arguments`（参数无效）、`room_not_found`（房间不存在）、`not_room_member`（未入房）和 `history_query_failed`（查询失败）。

[返回 README](../README.md)

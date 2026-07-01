TCP 收到字符串
        │
        ▼
msg = "to|Alice|hello"
        │
        ▼
protocol.Parse(msg)
        │
        ▼
cmd = Command{
    Type: CmdPrivate,
    Args: ["Alice", "hello"],
    Raw: "to|Alice|hello",
}
        │
        ▼
switch cmd.Type
        │
        ├── CmdWho      → handlerWho()
        ├── CmdRename   → handlerRename()
        └── CmdPrivate  → handlerPrivateChat()
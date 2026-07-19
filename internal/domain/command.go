package domain

type CommandType string

const (
	CmdUnknown CommandType = "unknonw"
	CmdWho     CommandType = "who"
	CmdRename  CommandType = "rename"
	CmdPrivate CommandType = "private"
	CmdRooms   CommandType = "rooms"
	CmdCreate  CommandType = "create"
	CmdJoin    CommandType = "join"
	CmdLeave   CommandType = "leave"
	CmdRoom    CommandType = "room"
	CmdWhere   CommandType = "where"
	CmdMembers CommandType = "members"
	CmdHelp    CommandType = "help"
	CmdPublic  CommandType = "public"
	CmdAuth    CommandType = "auth"
)

type Command struct {
	Type CommandType
	Args []string
	Raw  string
}

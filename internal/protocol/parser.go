package protocol

import (
	"strings"

	"IM-system/internal/domain"
)

func Parse(raw string) domain.Command {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.Command{Type: domain.CmdUnknown, Raw: raw}
	}

	parts := strings.Split(raw, "|")
	name := parts[0]
	args := parts[1:]

	switch name {
	case "who":
		return domain.Command{Type: domain.CmdWho, Args: args, Raw: raw}
	case "rename":
		return domain.Command{Type: domain.CmdRename, Args: args, Raw: raw}
	case "to":
		return domain.Command{Type: domain.CmdPrivate, Args: args, Raw: raw}
	case "rooms":
		return domain.Command{Type: domain.CmdRooms, Args: args, Raw: raw}
	case "create":
		return domain.Command{Type: domain.CmdCreate, Args: args, Raw: raw}
	case "join":
		return domain.Command{Type: domain.CmdJoin, Args: args, Raw: raw}
	case "leave":
		return domain.Command{Type: domain.CmdLeave, Args: args, Raw: raw}
	case "room":
		return domain.Command{Type: domain.CmdRoom, Args: args, Raw: raw}
	case "where":
		return domain.Command{Type: domain.CmdWhere, Args: args, Raw: raw}
	case "members":
		return domain.Command{Type: domain.CmdMembers, Args: args, Raw: raw}
	case "help":
		return domain.Command{Type: domain.CmdHelp, Args: args, Raw: raw}
	case "auth":
		return domain.Command{Type: domain.CmdAuth, Args: args, Raw: raw}
	default:
		return domain.Command{Type: domain.CmdPublic, Args: args, Raw: raw}
	}
}

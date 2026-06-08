package cloudflow

import "strings"

const (
	TypeResearch = "research"
	TypeSwarm    = "swarm"
	TypeAuto     = "auto"
)

type SlashCommand struct {
	Type     string
	Command  string
	Strategy string
	Query    string
}

func ParseSlash(text string) *SlashCommand {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return nil
	}
	rest := trimmed[1:]
	sp := strings.IndexByte(rest, ' ')
	if sp < 0 {
		return nil
	}
	cmd := strings.ToLower(rest[:sp])
	args := strings.TrimSpace(rest[sp+1:])
	if args == "" {
		return nil
	}
	switch cmd {
	case "research":
		strategy := "standard"
		query := args
		first, afterFirst, hasSpace := strings.Cut(args, " ")
		switch first {
		case "quick", "standard", "deep", "academic":
			if !hasSpace {
				return nil
			}
			strategy = first
			query = strings.TrimSpace(afterFirst)
		}
		if query == "" {
			return nil
		}
		return &SlashCommand{Type: TypeResearch, Command: "/research", Strategy: strategy, Query: query}
	case "swarm":
		return &SlashCommand{Type: TypeSwarm, Command: "/swarm", Query: args}
	case "dag":
		return &SlashCommand{Type: TypeAuto, Command: "/dag", Query: args}
	default:
		return nil
	}
}

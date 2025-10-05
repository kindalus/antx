package cli

import (
	prompt "github.com/c-bata/go-prompt"
)

// Command defines the interface for a CLI command.
type Command interface {
	Execute(args []string)
	Suggest(d prompt.Document) []prompt.Suggest
	GetName() string
	GetDescription() string
}

var commands = make(map[string]Command)

// RegisterCommand registers a new command.
func RegisterCommand(cmd Command) {
	commands[cmd.GetName()] = cmd
}

// GetCommands returns the registered commands.
func GetCommands() map[string]Command {
	return commands
}

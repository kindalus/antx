package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/kindalus/antx/antbox"

	prompt "github.com/c-bata/go-prompt"
)

var (
	client            antbox.Antbox
	currentFolder     = "--root--"
	currentFolderName = "root"
	currentNodes      []antbox.Node
)

func executor(in string) {
	in = strings.TrimSpace(in)
	parts := strings.Split(in, " ")
	commandName := parts[0]
	args := parts[1:]

	if cmd, ok := commands[commandName]; ok {
		cmd.Execute(args)
	} else {
		fmt.Println("Unknown command: " + commandName)
	}

	fmt.Println("")
}

func completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	args := strings.Split(d.TextBeforeCursor(), " ")
	commandName := args[0]

	if len(args) == 1 {
		var suggests []prompt.Suggest
		for name, cmd := range commands {
			if strings.HasPrefix(name, commandName) {
				suggests = append(suggests, prompt.Suggest{Text: name, Description: cmd.GetDescription()})
			}
		}
		return suggests
	}

	if cmd, ok := commands[commandName]; ok {
		return cmd.Suggest(d)
	}

	return []prompt.Suggest{}
}

func Start(serverURL, apiKey, root, jwt string, debug bool) {
	client = antbox.NewClient(serverURL, apiKey, root, jwt, debug)
	if root != "" {
		if err := client.Login(); err != nil {
			fmt.Println("Login failed:", err)
			os.Exit(1)
		}
	}

	// Initial ls
	if cmd, ok := commands["ls"]; ok {
		cmd.Execute([]string{})
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("Antbox CLI"),
		prompt.OptionCompletionOnDown(),
		prompt.OptionLivePrefix(func() (string, bool) {
			return fmt.Sprintf("%s >>> ", currentFolderName), true
		}),
	)
	p.Run()
}

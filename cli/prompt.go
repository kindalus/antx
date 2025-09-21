package cli

import (
	"antbox-cli/antbox"
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

var (
	client        antbox.Antbox
	currentFolder = "--root--"
	currentNodes  []antbox.Node
)

func executor(in string) {
	in = strings.TrimSpace(in)
	parts := strings.Split(in, " ")
	command := parts[0]
	args := parts[1:]

	switch command {
	case "ls":
		ls(args)
	case "pwd":
		pwd()
	case "stat":
		stat(args)
	case "cd":
		cd(args)
	case "mkdir":
		mkdir(args)
	case "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		fmt.Println("Unknown command: " + command)
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	args := strings.Split(in.Text, " ")

	if len(args) <= 1 {
		word := strings.TrimSpace(in.Text)

		// Only show suggestions if user has typed at least 2 characters
		if len(word) < 2 {
			return []prompt.Suggest{}
		}

		suggests := []prompt.Suggest{
			{Text: "ls", Description: "List content of a folder"},
			{Text: "pwd", Description: "Show current location"},
			{Text: "stat", Description: "Show node properties"},
			{Text: "cd", Description: "Change directory"},
			{Text: "mkdir", Description: "Create a directory"},
			{Text: "exit", Description: "Exit the CLI"},
		}

		var filtered []prompt.Suggest
		for _, suggest := range suggests {
			if strings.HasPrefix(strings.ToLower(suggest.Text), strings.ToLower(word)) {
				filtered = append(filtered, suggest)
			}
		}
		return filtered
	}

	command := args[0]
	if command == "ls" || command == "stat" || command == "cd" {
		// For command arguments, we need to get the current argument being typed
		var word string
		if len(args) > 1 {
			word = args[len(args)-1]
		} else {
			word = ""
		}

		// Only show suggestions if user has typed at least 2 characters for node names
		if len(word) < 2 {
			return []prompt.Suggest{}
		}

		var nodeSuggestions []prompt.Suggest
		addedUUIDs := make(map[string]bool) // Track added UUIDs to avoid duplicates

		for _, node := range currentNodes {
			if command == "cd" {
				// For cd command, only suggest folders and always use UUID
				if node.Mimetype == "application/vnd.antbox.folder" {
					if (strings.HasPrefix(strings.ToLower(node.Title), strings.ToLower(word)) ||
						strings.HasPrefix(strings.ToLower(node.UUID), strings.ToLower(word))) &&
						!addedUUIDs[node.UUID] {
						nodeSuggestions = append(nodeSuggestions, prompt.Suggest{
							Text:        node.UUID,
							Description: "📁 " + node.Title,
						})
						addedUUIDs[node.UUID] = true
					}
				}
			} else {
				// For other commands, allow both UUID and title matching for all node types
				if strings.HasPrefix(strings.ToLower(node.UUID), strings.ToLower(word)) {
					nodeSuggestions = append(nodeSuggestions, prompt.Suggest{Text: node.UUID, Description: node.Title})
				}
				if strings.HasPrefix(strings.ToLower(node.Title), strings.ToLower(word)) {
					nodeSuggestions = append(nodeSuggestions, prompt.Suggest{Text: node.Title, Description: node.UUID})
				}
			}
		}
		return nodeSuggestions
	}

	return []prompt.Suggest{}
}

func Start(serverURL, apiKey, root, jwt string) {
	client = antbox.NewClient(serverURL, apiKey, root, jwt)
	if root != "" {
		if err := client.Login(); err != nil {
			fmt.Println("Login failed:", err)
			os.Exit(1)
		}
	}

	ls([]string{})

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("Antbox CLI"),
		prompt.OptionCompletionOnDown(),
	)
	p.Run()
}

func ls(args []string) {
	var folder string
	if len(args) > 0 {
		folder = args[0]
	} else {
		folder = currentFolder
	}

	nodes, err := client.ListNodes(folder)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	currentNodes = nodes

	for _, node := range nodes {
		fmt.Printf("%s\t%s\t%s\n", node.UUID, node.Title, node.Mimetype)
	}
}

func pwd() {
	fmt.Println(currentFolder)
}

func stat(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: stat <uuid>")
		return
	}

	node, err := client.GetNode(args[0])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("UUID: %s\n", node.UUID)
	fmt.Printf("Title: %s\n", node.Title)
	fmt.Printf("Mimetype: %s\n", node.Mimetype)
	fmt.Printf("Parent: %s\n", node.Parent)
	fmt.Printf("Owner: %s\n", node.Owner)
	fmt.Printf("Group: %s\n", node.Group)
	fmt.Printf("Size: %d\n", node.Size)
	fmt.Printf("Created At: %s\n", node.CreatedAt)
	fmt.Printf("Modified At: %s\n", node.ModifiedAt)
}

func cd(args []string) {
	if len(args) == 0 {
		currentFolder = "--root--"
	} else if args[0] == ".." {
		if currentFolder == "--root--" {
			return
		}
		node, err := client.GetNode(currentFolder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		currentFolder = node.Parent
	} else {
		currentFolder = args[0]
	}
	ls([]string{})
}

func mkdir(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: mkdir <name>")
		return
	}

	fn := strings.Join(args, " ")

	// Remove initial and final ' or " if present
	if (strings.HasPrefix(fn, "'") && strings.HasSuffix(fn, "'")) || (strings.HasPrefix(fn, "\"") && strings.HasSuffix(fn, "\"")) {
		fn = fn[1 : len(fn)-1]
	}

	_, err := client.CreateFolder(currentFolder, fn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ls([]string{})
}

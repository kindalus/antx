package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	markdown "go.xrstf.de/go-term-markdown"
)

type RagCommand struct {
}

func (c *RagCommand) GetName() string {
	return "rag"
}

func (c *RagCommand) GetDescription() string {
	return "Send message to RAG agent"
}

func (c *RagCommand) Execute(args []string) {
	var useLocation bool
	var messageArgs []string

	// Parse flags and arguments
	i := 0
	stayInLoop := true
	for i < len(args) && stayInLoop {
		switch args[i] {
		case "-l":
			useLocation = true
			i++
		default:
			// Remaining arguments are the message
			messageArgs = args[i:]
			stayInLoop = false
		}
	}

	// If no message provided, enter interactive mode
	if len(messageArgs) == 0 {
		c.startInteractiveSession(useLocation)
		return
	}

	// Single message mode
	message := strings.Join(messageArgs, " ")
	c.sendMessage(message, nil, useLocation)
}

func (c *RagCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	lastArg := args[len(args)-1]

	// Suggest flags if we're typing a flag or at the beginning
	if strings.HasPrefix(lastArg, "-") || (len(args) <= 3 && !strings.HasSuffix(text, " ")) {
		var suggestions []prompt.Suggest
		if !strings.Contains(text, "-l") {
			suggestions = append(suggestions, prompt.Suggest{Text: "-l", Description: "Use current location as context"})
		}
		return suggestions
	}

	return []prompt.Suggest{}
}

// startInteractiveSession starts an interactive RAG session
func (c *RagCommand) startInteractiveSession(useLocation bool) {
	fmt.Println("Starting interactive RAG session")
	if useLocation {
		fmt.Printf("Using location context: %s\n", currentFolderName)
	}
	fmt.Println("Type 'exit' or press Ctrl+D to exit the session.")
	fmt.Println()

	// Create interactive session context
	sessionContext := &RagSessionContext{
		useLocation: useLocation,
		command:     c,
	}

	// Create a new prompt for the RAG session
	p := prompt.New(
		sessionContext.executeMessage,
		func(d prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
		prompt.OptionTitle("RAG Session"),
		prompt.OptionPrefix("You: "),
	)
	p.Run()
}

// RagSessionContext holds the context for an interactive RAG session
type RagSessionContext struct {
	useLocation bool
	command     *RagCommand
	history     []map[string]any
}

func (ctx *RagSessionContext) executeMessage(input string) {
	input = strings.TrimSpace(input)

	// Check for exit command
	if input == "exit" {
		fmt.Println("Exiting RAG session...")
		return
	}

	// Skip empty messages
	if input == "" {
		return
	}

	// Send message and display response
	history := ctx.command.sendMessage(input, ctx.history, ctx.useLocation)

	if history != nil {
		ctx.history = history
	}
}

// sendMessage sends a single message to the RAG agent and displays the response
func (c *RagCommand) sendMessage(message string, history []map[string]any, useLocation bool) []map[string]any {
	options := make(map[string]any)

	if useLocation {
		options["parent"] = currentFolder
	}

	if history != nil {
		options["history"] = history
	}

	response, err := client.RagChat(message, options)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	result := markdown.Render(response.Text, 100, 11)

	fmt.Printf("\033[32mAssistant:\033[0m %s\n", strings.Trim(string(result), " "))
	return response.History
}

func init() {
	RegisterCommand(&RagCommand{})
}

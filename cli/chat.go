package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
)

type ChatCommand struct{}

func (c *ChatCommand) GetName() string {
	return "chat"
}

func (c *ChatCommand) GetDescription() string {
	return "Send message to specific agent"
}

func (c *ChatCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: chat [options] <agent_uuid> [message]")
		fmt.Println("Options:")
		fmt.Println("  -t <temperature>  Temperature for response generation (0.0-1.0)")
		fmt.Println("  -m <max_tokens>   Maximum tokens in the response")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  agent_uuid: UUID of the agent to chat with")
		fmt.Println("  message: Initial message to send (optional)")
		fmt.Println()
		fmt.Println("Interactive mode:")
		fmt.Println("  Chat is always interactive. If a message is provided, it's sent first.")
		fmt.Println("  Type 'exit' or press Ctrl+D to exit the session.")
		return
	}

	var temperature *float64
	var maxTokens *int
	var agentUUID string
	var messageArgs []string

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		switch args[i] {
		case "-t":
			if i+1 >= len(args) {
				fmt.Println("Error: -t requires a temperature value")
				return
			}
			temp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || temp < 0 || temp > 1 {
				fmt.Println("Error: Temperature must be a number between 0.0 and 1.0")
				return
			}
			temperature = &temp
			i += 2
		case "-m":
			if i+1 >= len(args) {
				fmt.Println("Error: -m requires a max tokens value")
				return
			}
			tokens, err := strconv.Atoi(args[i+1])
			if err != nil || tokens <= 0 {
				fmt.Println("Error: Max tokens must be a positive integer")
				return
			}
			maxTokens = &tokens
			i += 2

		default:
			// First non-option argument is agent UUID
			if agentUUID == "" {
				agentUUID = args[i]
				messageArgs = args[i+1:]
				goto parseComplete
			}
			i++
		}
	}

parseComplete:

	if agentUUID == "" {
		fmt.Println("Error: Agent UUID is required")
		return
	}

	// Always enter interactive mode, optionally with initial message
	var initialMessage string
	if len(messageArgs) > 0 {
		initialMessage = strings.Join(messageArgs, " ")
	}

	c.startInteractiveSession(agentUUID, initialMessage, temperature, maxTokens)
}

func (c *ChatCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	lastArg := args[len(args)-1]

	// Count actual arguments (excluding the command name and flags)
	argCount := 0
	flagCount := 0
	for i := 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagCount++
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagCount++ // Skip flag value
				i++
			}
		} else if flagCount == 0 || i > flagCount {
			argCount++
		}
	}

	// Adjust for current typing
	if !strings.HasSuffix(text, " ") && len(args) > 1 {
		if strings.HasPrefix(lastArg, "-") {
			// Currently typing a flag
		} else if argCount > 0 {
			argCount-- // Still typing the current argument
		}
	}

	// Suggest flags if we're typing a flag
	if strings.HasPrefix(lastArg, "-") {
		return []prompt.Suggest{
			{Text: "-t", Description: "Set temperature (0.0-1.0)"},
			{Text: "-m", Description: "Set max tokens"},
		}
	}

	// Suggest agent UUID if we haven't specified one yet
	if argCount == 0 {
		// Use cached agents
		agents := GetCachedAgents()
		var suggests []prompt.Suggest
		currentWord := d.GetWordBeforeCursor()
		for _, agent := range agents {
			if strings.HasPrefix(strings.ToLower(agent.UUID), strings.ToLower(currentWord)) ||
				strings.HasPrefix(strings.ToLower(agent.Title), strings.ToLower(currentWord)) {
				suggests = append(suggests, prompt.Suggest{
					Text:        agent.UUID,
					Description: agent.Title,
				})
			}
		}
		return suggests
	}

	return []prompt.Suggest{}
}

// startInteractiveSession starts an interactive chat session with the specified agent
func (c *ChatCommand) startInteractiveSession(agentUUID string, initialMessage string, temperature *float64, maxTokens *int) {
	// Find agent name for display
	agentName := agentUUID
	for _, agent := range GetCachedAgents() {
		if agent.UUID == agentUUID {
			agentName = agent.Title
			break
		}
	}

	fmt.Printf("Starting interactive chat with %s\n", agentName)
	fmt.Println("Type 'exit' or press Ctrl+D to exit the session.")
	fmt.Println()

	// Send initial message if provided
	if initialMessage != "" {
		fmt.Printf("You: %s\n", initialMessage)
		c.sendMessage(agentUUID, initialMessage, temperature, maxTokens)
		fmt.Println()
	}

	// Create interactive session context
	sessionContext := &ChatSessionContext{
		agentUUID:   agentUUID,
		temperature: temperature,
		maxTokens:   maxTokens,
		command:     c,
	}

	// Create a new prompt for the chat session
	p := prompt.New(
		sessionContext.executeMessage,
		func(d prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
		prompt.OptionTitle(fmt.Sprintf("Chat with %s", agentName)),
		prompt.OptionPrefix("You: "),
	)
	p.Run()
}

// ChatSessionContext holds the context for an interactive chat session
type ChatSessionContext struct {
	agentUUID   string
	temperature *float64
	maxTokens   *int
	command     *ChatCommand
}

func (ctx *ChatSessionContext) executeMessage(input string) {
	input = strings.TrimSpace(input)

	// Check for exit command
	if input == "exit" {
		fmt.Println("Exiting chat session...")
		return
	}

	// Skip empty messages
	if input == "" {
		return
	}

	// Send message and display response
	ctx.command.sendMessage(ctx.agentUUID, input, ctx.temperature, ctx.maxTokens)
}

// sendMessage sends a single message to the agent and displays the response
func (c *ChatCommand) sendMessage(agentUUID string, message string, temperature *float64, maxTokens *int) {
	chatHistory, err := client.ChatWithAgent(agentUUID, message, "", temperature, maxTokens, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Find the last model response from the chat history
	for i := len(chatHistory) - 1; i >= 0; i-- {
		msg := chatHistory[i]
		if msg.Role == "model" {
			for _, part := range msg.Parts {
				if part.Text != nil {
					fmt.Printf("Assistant: %s\n", *part.Text)
					return
				}
			}
		}
	}

	// If no model response found, show that no response was received
	fmt.Println("Assistant: (no response)")
}

func init() {
	RegisterCommand(&ChatCommand{})
}

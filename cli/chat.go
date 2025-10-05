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
	if len(args) < 2 {
		fmt.Println("Usage: chat [options] <agent_uuid> <message>")
		fmt.Println("Options:")
		fmt.Println("  -t <temperature>  Temperature for response generation (0.0-1.0)")
		fmt.Println("  -m <max_tokens>   Maximum tokens in the response")
		fmt.Println("  -c <conversation_id>  Conversation ID to continue previous conversation")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  agent_uuid: UUID of the agent to chat with")
		fmt.Println("  message: Message to send to the agent")
		return
	}

	var temperature *float64
	var maxTokens *int
	var conversationID string
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
		case "-c":
			if i+1 >= len(args) {
				fmt.Println("Error: -c requires a conversation ID")
				return
			}
			conversationID = args[i+1]
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

	if len(messageArgs) == 0 {
		fmt.Println("Error: Message is required")
		return
	}

	message := strings.Join(messageArgs, " ")

	response, err := client.ChatWithAgent(agentUUID, message, conversationID, temperature, maxTokens)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(response)
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
			{Text: "-c", Description: "Set conversation ID"},
		}
	}

	// Suggest agent UUID if we haven't specified one yet
	if argCount == 0 {
		agents, err := client.ListAgents()
		if err != nil {
			return []prompt.Suggest{
				{Text: "", Description: "Enter agent UUID"},
			}
		}

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

func init() {
	RegisterCommand(&ChatCommand{})
}

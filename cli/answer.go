package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
)

type AnswerCommand struct{}

func (c *AnswerCommand) GetName() string {
	return "answer"
}

func (c *AnswerCommand) GetDescription() string {
	return "Send question to specific agent for answering"
}

func (c *AnswerCommand) Execute(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: answer [options] <agent_uuid> <question>")
		fmt.Println("Options:")
		fmt.Println("  -t <temperature>  Temperature for response generation (0.0-1.0)")
		fmt.Println("  -m <max_tokens>   Maximum tokens in the response")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  agent_uuid: UUID of the agent to ask")
		fmt.Println("  question: Question to ask the agent")
		return
	}

	var temperature *float64
	var maxTokens *int
	var agentUUID string
	var questionArgs []string

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
				questionArgs = args[i+1:]
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

	if len(questionArgs) == 0 {
		fmt.Println("Error: Question is required")
		return
	}

	question := strings.Join(questionArgs, " ")

	// Find agent name for display
	agentName := agentUUID
	for _, agent := range GetCachedAgents() {
		if agent.UUID == agentUUID {
			agentName = agent.Title
			break
		}
	}

	// Show loading animation while waiting for response
	animation := StartLoadingAnimationWithStyle(fmt.Sprintf("Asking %s", agentName), SpinnerStyle)
	chatHistory, err := client.AnswerFromAgent(agentUUID, question, temperature, maxTokens)

	if err != nil {
		animation.StopWithMessage(fmt.Sprintf("✗ Error asking %s", agentName))
		fmt.Println("Error:", err)
		return
	}

	animation.StopWithMessage(fmt.Sprintf("✓ Response from %s:", agentName))

	// Find the last model response from the chat history
	for i := len(chatHistory) - 1; i >= 0; i-- {
		msg := chatHistory[i]
		if msg.Role == "model" {
			for _, part := range msg.Parts {
				if part.Text != nil {
					fmt.Println(*part.Text)
					return
				}
			}
		}
	}

	// If no model response found, show that no response was received
	fmt.Println("(no response)")
}

func (c *AnswerCommand) Suggest(d prompt.Document) []prompt.Suggest {
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

func init() {
	RegisterCommand(&AnswerCommand{})
}

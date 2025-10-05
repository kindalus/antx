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

	response, err := client.AnswerFromAgent(agentUUID, question, temperature, maxTokens)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(response)
}

func (c *AnswerCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	lastArg := args[len(args)-1]

	// Suggest flags if we're typing a flag or at the beginning
	if strings.HasPrefix(lastArg, "-") || (len(args) <= 3 && !strings.HasSuffix(text, " ")) {
		return []prompt.Suggest{
			{Text: "-t", Description: "Set temperature (0.0-1.0)"},
			{Text: "-m", Description: "Set max tokens"},
		}
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&AnswerCommand{})
}

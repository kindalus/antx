package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
)

type RagCommand struct{}

func (c *RagCommand) GetName() string {
	return "rag"
}

func (c *RagCommand) GetDescription() string {
	return "Send message to RAG agent"
}

func (c *RagCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: rag [options] <message>")
		fmt.Println("Options:")
		fmt.Println("  -l                Use current location as parent context")
		fmt.Println("  -c <conversation_id>  Conversation ID to continue previous conversation")
		fmt.Println("  -f <field>=<value>    Add custom filter (can be used multiple times)")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  message: Message for RAG chat")
		return
	}

	var useLocation bool
	var conversationID string
	var customFilters map[string]interface{}
	var messageArgs []string

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		switch args[i] {
		case "-l":
			useLocation = true
			i++
		case "-c":
			if i+1 >= len(args) {
				fmt.Println("Error: -c requires a conversation ID")
				return
			}
			conversationID = args[i+1]
			i += 2
		case "-f":
			if i+1 >= len(args) {
				fmt.Println("Error: -f requires a filter in format field=value")
				return
			}
			filterParts := strings.SplitN(args[i+1], "=", 2)
			if len(filterParts) != 2 {
				fmt.Println("Error: Filter must be in format field=value")
				return
			}
			if customFilters == nil {
				customFilters = make(map[string]interface{})
			}
			// Try to parse as number first, then boolean, then string
			value := filterParts[1]
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				customFilters[filterParts[0]] = num
			} else if bool, err := strconv.ParseBool(value); err == nil {
				customFilters[filterParts[0]] = bool
			} else {
				customFilters[filterParts[0]] = value
			}
			i += 2
		default:
			// Remaining arguments are the message
			messageArgs = args[i:]
			goto parseComplete
		}
	}

parseComplete:

	if len(messageArgs) == 0 {
		fmt.Println("Error: No message provided")
		return
	}

	message := strings.Join(messageArgs, " ")

	// Get or create session if conversation ID is provided
	var session *Session
	if conversationID != "" {
		session = GetOrCreateSession(conversationID)

		// Add user message to session history
		session.AddMessage("user", message)
	}

	var filters map[string]interface{}
	if useLocation || customFilters != nil {
		filters = make(map[string]interface{})
		if useLocation {
			filters["parent"] = currentFolder
		}
		// Add custom filters
		for k, v := range customFilters {
			filters[k] = v
		}
	}

	// Get conversation history if this is a continuing conversation
	var history []map[string]interface{}
	if session != nil && !session.IsEmpty() {
		// For subsequent messages, include history (excluding the current user message we just added)
		allHistory := session.GetHistoryAsMap()
		if len(allHistory) > 1 {
			// Exclude the last message (current user message) from history sent to API
			history = allHistory[:len(allHistory)-1]
		}
	}

	response, err := client.RagChat(message, conversationID, filters, history)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Add assistant response to session history
	if session != nil {
		session.AddMessage("assistant", response)
	}

	fmt.Println(response)
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
		if !strings.Contains(text, "-c") {
			suggestions = append(suggestions, prompt.Suggest{Text: "-c", Description: "Set conversation ID"})
		}
		if !strings.Contains(text, "-f") {
			suggestions = append(suggestions, prompt.Suggest{Text: "-f", Description: "Add custom filter (field=value)"})
		}
		return suggestions
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&RagCommand{})
}

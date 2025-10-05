package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

type RunCommand struct{}

func (c *RunCommand) GetName() string {
	return "run"
}

func (c *RunCommand) GetDescription() string {
	return "Run an action on a node with optional parameters"
}

func (c *RunCommand) Execute(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: run <action_uuid> <node_uuid> [param=value...]")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  action_uuid: UUID of the action to run")
		fmt.Println("  node_uuid: UUID of the node to run the action on")
		fmt.Println("  param=value: Optional parameters in key=value format")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  run abc123 def456")
		fmt.Println("  run abc123 def456 format=pdf quality=high")
		return
	}

	actionUUID := args[0]
	nodeUUID := args[1]

	// Parse parameters from remaining arguments
	parameters := make(map[string]interface{})
	for i := 2; i < len(args); i++ {
		param := args[i]
		if strings.Contains(param, "=") {
			parts := strings.SplitN(param, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Try to convert value to appropriate type
			if value == "true" {
				parameters[key] = true
			} else if value == "false" {
				parameters[key] = false
			} else {
				// Try to parse as number, otherwise keep as string
				if intVal, err := parseAsInt(value); err == nil {
					parameters[key] = intVal
				} else if floatVal, err := parseAsFloat(value); err == nil {
					parameters[key] = floatVal
				} else {
					parameters[key] = value
				}
			}
		} else {
			fmt.Printf("Warning: Ignoring invalid parameter format: %s (expected key=value)\n", param)
		}
	}

	// Create action run request
	request := antbox.ActionRunRequest{
		UUIDs:      []string{nodeUUID},
		Parameters: parameters,
	}

	// Execute the action
	result, err := client.RunAction(actionUUID, request)
	if err != nil {
		fmt.Println("Error running action:", err)
		return
	}

	// Display the result
	fmt.Println("Action executed successfully:")
	printResult(result)
}

func (c *RunCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	// Count actual arguments (excluding the command name)
	argCount := len(args) - 1
	if !strings.HasSuffix(text, " ") && len(args) > 1 {
		argCount = len(args) - 2 // We're still typing the current argument
	}

	switch argCount {
	case 0:
		// Suggesting action UUID - we could potentially list actions here
		return []prompt.Suggest{
			{Text: "", Description: "Enter action UUID"},
		}
	case 1:
		// Suggesting node UUID
		return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
	default:
		// Suggesting parameters - show common parameter examples
		currentWord := d.GetWordBeforeCursor()
		if !strings.Contains(currentWord, "=") {
			return []prompt.Suggest{
				{Text: "format=", Description: "Output format parameter"},
				{Text: "quality=", Description: "Quality parameter"},
				{Text: "size=", Description: "Size parameter"},
				{Text: "mode=", Description: "Mode parameter"},
			}
		}
	}

	return []prompt.Suggest{}
}

func parseAsInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseAsFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

func printResult(result map[string]interface{}) {
	for key, value := range result {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func init() {
	RegisterCommand(&RunCommand{})
}

package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type CallCommand struct{}

func (c *CallCommand) GetName() string {
	return "call"
}

func (c *CallCommand) GetDescription() string {
	return "Run an extension with optional parameters"
}

func (c *CallCommand) Execute(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: call <extension_uuid> [param=value...]")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  extension_uuid: UUID of the extension to run")
		fmt.Println("  param=value: Optional parameters in key=value format")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  call abc123")
		fmt.Println("  call abc123 input=hello format=json timeout=30")
		return
	}

	extensionUUID := args[0]

	// Parse parameters from remaining arguments
	parameters := make(map[string]interface{})
	for i := 1; i < len(args); i++ {
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

	// Execute the extension
	result, err := client.RunExtension(extensionUUID, parameters)
	if err != nil {
		fmt.Println("Error running extension:", err)
		return
	}

	// Display the result
	fmt.Println("Extension executed successfully:")
	printExtensionResult(result)
}

func (c *CallCommand) Suggest(d prompt.Document) []prompt.Suggest {
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
		// Suggesting extension UUID - use cached extensions
		extensions := GetCachedExtensions()
		var suggests []prompt.Suggest
		currentWord := d.GetWordBeforeCursor()
		for _, extension := range extensions {
			if strings.HasPrefix(strings.ToLower(extension.UUID), strings.ToLower(currentWord)) ||
				strings.HasPrefix(strings.ToLower(extension.Title), strings.ToLower(currentWord)) {
				suggests = append(suggests, prompt.Suggest{
					Text:        extension.UUID,
					Description: extension.Title,
				})
			}
		}
		return suggests
	default:
		// Suggesting parameters - show common parameter examples
		currentWord := d.GetWordBeforeCursor()
		if !strings.Contains(currentWord, "=") {
			return []prompt.Suggest{
				{Text: "input=", Description: "Input data parameter"},
				{Text: "format=", Description: "Output format parameter"},
				{Text: "timeout=", Description: "Timeout parameter"},
				{Text: "config=", Description: "Configuration parameter"},
				{Text: "mode=", Description: "Mode parameter"},
			}
		}
	}

	return []prompt.Suggest{}
}

func printExtensionResult(result interface{}) {
	switch v := result.(type) {
	case string:
		fmt.Printf("  %s\n", v)
	case map[string]interface{}:
		for key, value := range v {
			fmt.Printf("  %s: %v\n", key, value)
		}
	case []interface{}:
		for i, item := range v {
			fmt.Printf("  [%d]: %v\n", i, item)
		}
	default:
		fmt.Printf("  %v\n", v)
	}
}

func init() {
	RegisterCommand(&CallCommand{})
}

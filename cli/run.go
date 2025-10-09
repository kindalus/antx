package cli

import (
	"fmt"
	"strconv"
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
	parameters := make(map[string]any)
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
				if intVal, err := strconv.Atoi(value); err == nil {
					parameters[key] = intVal
				} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
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
		// Suggesting action UUID - filter only actions that can be run
		actions := GetCachedActions()
		var suggests []prompt.Suggest
		currentWord := d.GetWordBeforeCursor()
		for _, action := range actions {
			// Only suggest actions that are exposed as actions and can be run manually
			if !action.ExposeAsAction || !action.RunManually {
				continue
			}

			if strings.HasPrefix(strings.ToLower(action.UUID), strings.ToLower(currentWord)) ||
				strings.HasPrefix(strings.ToLower(action.Name), strings.ToLower(currentWord)) {
				description := action.Name
				if action.Description != "" {
					description = fmt.Sprintf("%s - %s", action.Name, action.Description)
				}
				suggests = append(suggests, prompt.Suggest{
					Text:        action.UUID,
					Description: description,
				})
			}
		}
		return suggests
	case 1:
		// Suggesting node UUID - filter based on the selected action's filters
		actionUUID := args[1] // The action UUID from first argument
		return c.getFilteredNodeSuggestions(d.GetWordBeforeCursor(), actionUUID)
	default:
		// Suggesting parameters based on the selected action's parameter definitions
		if len(args) >= 3 {
			actionUUID := args[1] // The action UUID from first argument
			return c.getActionParameterSuggestions(d.GetWordBeforeCursor(), actionUUID)
		}
		return []prompt.Suggest{}
	}
}

// getFilteredNodeSuggestions returns node suggestions filtered by the action's node filters
func (c *RunCommand) getFilteredNodeSuggestions(word, actionUUID string) []prompt.Suggest {
	// Find the action to get its filters
	actions := GetCachedActions()
	var selectedAction *antbox.Feature
	for _, action := range actions {
		if action.UUID == actionUUID {
			selectedAction = &action
			break
		}
	}

	// If action not found or has no filters, suggest all nodes
	if selectedAction == nil || selectedAction.Filters == nil {
		return getNodeSuggestions(word, nil)
	}

	// Create a filter function based on the action's filters
	filterFunc := func(node antbox.Node) bool {
		return c.nodeMatchesFilters(node, selectedAction.Filters)
	}

	return getNodeSuggestions(word, filterFunc)
}

// getActionParameterSuggestions returns parameter suggestions based on the action's parameter definitions
func (c *RunCommand) getActionParameterSuggestions(word, actionUUID string) []prompt.Suggest {
	currentWord := strings.TrimSpace(word)

	// If we're not typing a parameter (no = sign), suggest parameter names
	if !strings.Contains(currentWord, "=") {
		// Find the action to get its parameters
		actions := GetCachedActions()
		var selectedAction *antbox.Feature
		for _, action := range actions {
			if action.UUID == actionUUID {
				selectedAction = &action
				break
			}
		}

		var suggests []prompt.Suggest
		if selectedAction != nil && len(selectedAction.Parameters) > 0 {
			// Suggest actual action parameters
			for _, param := range selectedAction.Parameters {
				if strings.HasPrefix(strings.ToLower(param.Name), strings.ToLower(currentWord)) {
					description := fmt.Sprintf("%s (%s)", param.Description, param.Type)
					if param.Required {
						description += " - Required"
					}
					if param.DefaultValue != nil {
						description += fmt.Sprintf(" - Default: %v", param.DefaultValue)
					}
					suggests = append(suggests, prompt.Suggest{
						Text:        param.Name + "=",
						Description: description,
					})
				}
			}
		} else {
			// Fallback to common parameter names if no action found or no parameters defined
			commonParams := []struct {
				name        string
				description string
			}{
				{"format", "Output format parameter"},
				{"quality", "Quality parameter"},
				{"size", "Size parameter"},
				{"mode", "Mode parameter"},
			}

			for _, param := range commonParams {
				if strings.HasPrefix(strings.ToLower(param.name), strings.ToLower(currentWord)) {
					suggests = append(suggests, prompt.Suggest{
						Text:        param.name + "=",
						Description: param.description,
					})
				}
			}
		}
		return suggests
	}

	// If typing a parameter value (has = sign), could suggest values based on parameter type
	// For now, return empty suggestions for values
	return []prompt.Suggest{}
}

// nodeMatchesFilters checks if a node matches the given filters
func (c *RunCommand) nodeMatchesFilters(node antbox.Node, filters antbox.NodeFilters) bool {
	// This is a simplified filter matching implementation
	// In a real implementation, you'd need to properly parse and evaluate the filters
	// For now, we'll do basic matching on common filter types

	switch f := filters.(type) {
	case []interface{}:
		// Handle 1D array of filters (AND logic)
		for _, filter := range f {
			if filterArray, ok := filter.([]interface{}); ok && len(filterArray) >= 3 {
				field, _ := filterArray[0].(string)
				operator, _ := filterArray[1].(string)
				value := filterArray[2]

				if !c.evaluateFilter(node, field, operator, value) {
					return false
				}
			}
		}
		return true
	case [][]interface{}:
		// Handle 2D array of filters (OR of ANDs)
		for _, filterGroup := range f {
			allMatch := true
			for _, filter := range filterGroup {
				if filterArray, ok := filter.([]interface{}); ok && len(filterArray) >= 3 {
					field, _ := filterArray[0].(string)
					operator, _ := filterArray[1].(string)
					value := filterArray[2]

					if !c.evaluateFilter(node, field, operator, value) {
						allMatch = false
						break
					}
				}
			}
			if allMatch {
				return true // At least one group matched completely
			}
		}
		return false
	default:
		// Unknown filter format, allow all nodes
		return true
	}
}

// evaluateFilter evaluates a single filter condition against a node
func (c *RunCommand) evaluateFilter(node antbox.Node, field, operator string, value interface{}) bool {
	var nodeValue interface{}

	// Extract the field value from the node
	switch strings.ToLower(field) {
	case "title":
		nodeValue = node.Title
	case "mimetype":
		nodeValue = node.Mimetype
	case "size":
		nodeValue = node.Size
	case "owner":
		nodeValue = node.Owner
	default:
		// Unknown field, assume it doesn't match
		return false
	}

	// Evaluate based on operator
	switch strings.ToLower(operator) {
	case "=", "==", "eq", "equals":
		return fmt.Sprintf("%v", nodeValue) == fmt.Sprintf("%v", value)
	case "!=", "ne", "not_equals":
		return fmt.Sprintf("%v", nodeValue) != fmt.Sprintf("%v", value)
	case "contains", "match":
		nodeStr := strings.ToLower(fmt.Sprintf("%v", nodeValue))
		valueStr := strings.ToLower(fmt.Sprintf("%v", value))
		return strings.Contains(nodeStr, valueStr)
	case "starts_with", "startswith":
		nodeStr := strings.ToLower(fmt.Sprintf("%v", nodeValue))
		valueStr := strings.ToLower(fmt.Sprintf("%v", value))
		return strings.HasPrefix(nodeStr, valueStr)
	case "ends_with", "endswith":
		nodeStr := strings.ToLower(fmt.Sprintf("%v", nodeValue))
		valueStr := strings.ToLower(fmt.Sprintf("%v", value))
		return strings.HasSuffix(nodeStr, valueStr)
	case ">", "gt":
		if nodeNum, nodeErr := strconv.ParseFloat(fmt.Sprintf("%v", nodeValue), 64); nodeErr == nil {
			if valueNum, valueErr := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); valueErr == nil {
				return nodeNum > valueNum
			}
		}
		return false
	case "<", "lt":
		if nodeNum, nodeErr := strconv.ParseFloat(fmt.Sprintf("%v", nodeValue), 64); nodeErr == nil {
			if valueNum, valueErr := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); valueErr == nil {
				return nodeNum < valueNum
			}
		}
		return false
	case ">=", "gte":
		if nodeNum, nodeErr := strconv.ParseFloat(fmt.Sprintf("%v", nodeValue), 64); nodeErr == nil {
			if valueNum, valueErr := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); valueErr == nil {
				return nodeNum >= valueNum
			}
		}
		return false
	case "<=", "lte":
		if nodeNum, nodeErr := strconv.ParseFloat(fmt.Sprintf("%v", nodeValue), 64); nodeErr == nil {
			if valueNum, valueErr := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); valueErr == nil {
				return nodeNum <= valueNum
			}
		}
		return false
	default:
		// Unknown operator, assume it doesn't match
		return false
	}
}

func printResult(result map[string]any) {
	for key, value := range result {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func init() {
	RegisterCommand(&RunCommand{})
}

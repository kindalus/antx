package cli

import (
	"slices"
	"strconv"
	"strings"
)

func extractSingleFilter(searchText string) [][]any {
	tokens := strings.Fields(searchText)

	operators := []string{
		"==",
		"<=",
		">=",
		"<",
		">",
		"!=",
		"in",
		"not-in",
		"match",
		"contains",
		"contains-all",
		"contains-any",
		"not-contains",
		"contains-none",
		"~=",
	}

	if len(tokens) >= 2 && slices.Contains(operators, tokens[1]) {
		return [][]any{{tokens[0], tokens[1], strings.Join(tokens[2:], " ")}}
	}

	return [][]any{{":content", "~=", searchText}}
}

func convertValue(valueStr string) any {
	// Try to convert to number first
	if val, err := strconv.Atoi(valueStr); err == nil {
		return val
	}

	// Try to convert to float
	if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return val
	}

	// Try to convert to boolean
	if val, err := strconv.ParseBool(valueStr); err == nil {
		return val
	}

	// Return as string if no conversion worked
	return valueStr
}

// normalizeOperators adds spaces around operators in query strings
func normalizeOperators(input string) string {
	operators := []string{"==", ">=", "<=", "!=", "~=", ">", "<"}

	result := input
	for _, op := range operators {
		// Replace operator without spaces with operator with spaces
		spaced := " " + op + " "
		result = strings.ReplaceAll(result, op, spaced)
	}

	// Clean up multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

// mksmart creates a smart folder with the given arguments
func mksmart(args []string) {
	if cmd, ok := commands["mksmart"]; ok {
		cmd.Execute(args)
	}
}

// cd executes the cd command with the given arguments
func cd(args []string) {
	if cmd, ok := commands["cd"]; ok {
		cmd.Execute(args)
	}
}

// ls executes the ls command with the given arguments
func ls(args []string) {
	if cmd, ok := commands["ls"]; ok {
		cmd.Execute(args)
	}
}

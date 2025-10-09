package cli

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kindalus/antx/antbox"
)

func extractSingleFilter(searchText string) antbox.NodeFilters1D {
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
		return antbox.NodeFilters1D{
			antbox.NodeFilter{tokens[0], antbox.FilterOperator(tokens[1]), strings.Join(tokens[2:], " ")},
		}
	}

	return antbox.NodeFilters1D{
		antbox.NodeFilter{":content", antbox.FilterOperatorMatch, searchText},
	}
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
	// Process longer operators first to avoid substring replacement issues
	operators := []string{"==", ">=", "<=", "!=", "~=", ">", "<"}

	result := input
	for _, op := range operators {
		// Only replace operators that are not part of larger operators
		// by checking that they are either at boundaries or surrounded by non-operator chars
		var newResult strings.Builder
		i := 0
		for i < len(result) {
			if i <= len(result)-len(op) && result[i:i+len(op)] == op {
				// Found the operator, check if it's part of a larger operator
				isPartOfLarger := false

				// Check if this operator is part of a longer operator
				for _, longerOp := range operators {
					if len(longerOp) > len(op) {
						// Check if the current position is within a longer operator
						for j := max(0, i-len(longerOp)+len(op)); j <= min(i, len(result)-len(longerOp)); j++ {
							if j+len(longerOp) <= len(result) && result[j:j+len(longerOp)] == longerOp {
								if j <= i && i+len(op) <= j+len(longerOp) {
									isPartOfLarger = true
									break
								}
							}
						}
						if isPartOfLarger {
							break
						}
					}
				}

				if !isPartOfLarger {
					// Check if already properly spaced
					hasSpaceBefore := i == 0 || result[i-1] == ' '
					hasSpaceAfter := i+len(op) >= len(result) || result[i+len(op)] == ' '

					if !hasSpaceBefore {
						newResult.WriteString(" ")
					}
					newResult.WriteString(op)
					if !hasSpaceAfter {
						newResult.WriteString(" ")
					}
				} else {
					newResult.WriteString(op)
				}
				i += len(op)
			} else {
				newResult.WriteByte(result[i])
				i++
			}
		}

		result = newResult.String()
	}

	// Clean up multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// sortNodesForListing sorts nodes with directories first, then files, both alphabetically by title
func sortNodesForListing(nodes []antbox.Node) []antbox.Node {
	if len(nodes) == 0 {
		return nodes
	}

	// Separate directories and files
	var directories []antbox.Node
	var files []antbox.Node

	for _, node := range nodes {
		isFolder := node.Mimetype == "application/vnd.antbox.folder" ||
			node.Mimetype == "application/vnd.antbox.smartfolder"

		if isFolder {
			directories = append(directories, node)
		} else {
			files = append(files, node)
		}
	}

	// Sort directories alphabetically by title
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Title < directories[j].Title
	})

	// Sort files alphabetically by title
	sort.Slice(files, func(i, j int) bool {
		return files[i].Title < files[j].Title
	})

	// Combine directories first, then files
	result := make([]antbox.Node, 0, len(nodes))
	result = append(result, directories...)
	result = append(result, files...)

	return result
}

// formatModifiedDate formats a date string to local time with conditional year display
// Format: "mmm dd HH:mm" for current year, "mmm dd  yyyy" for other years
func formatModifiedDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}

	// Parse the ISO 8601 date string
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "N/A"
	}

	// Convert to local time
	localTime := parsedTime.Local()
	currentYear := time.Now().Year()

	// Format based on year
	if localTime.Year() == currentYear {
		return localTime.Format("Jan 02 15:04")
	} else {
		return localTime.Format("Jan 02  2006")
	}
}

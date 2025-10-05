package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type FindCommand struct{}

func (c *FindCommand) GetName() string {
	return "find"
}

func (c *FindCommand) GetDescription() string {
	return "Find nodes using filter criteria"
}

func (c *FindCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: find <criteria>")
		fmt.Println("  Simple: find some text")
		fmt.Println("  Complex: find title == Document,owner ~= admin,size > 1000")
		return
	}

	searchText := strings.Join(args, " ")
	var filters any

	if !strings.Contains(searchText, ",") {
		// Simple search: use title field for text search
		filters = extractSingleFilter(searchText)
	} else {
		// Complex search: parse comma-separated criteria
		var filterList [][]any
		parts := strings.Split(searchText, ",")

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Split by whitespace to get field, operator, value
			tokens := strings.Fields(part)

			if len(tokens) < 3 {
				// Skip incomplete filters
				continue
			} else if len(tokens) == 3 {
				// Exact match: field operator value
				value := convertValue(tokens[2])
				filter := []any{tokens[0], tokens[1], value}
				filterList = append(filterList, filter)
			} else {
				// More than 3 tokens: field operator "rest as value"
				valueStr := strings.Join(tokens[2:], " ")
				value := convertValue(valueStr)
				filter := []any{tokens[0], tokens[1], value}
				filterList = append(filterList, filter)
			}
		}

		if len(filterList) == 0 {
			fmt.Println("No valid filters found in criteria")
			return
		}

		filters = filterList
	}

	fmt.Printf("Debug: Sending filters to API: %+v\n", filters)
	result, err := client.FindNodes(filters, 20, 1)
	if err != nil {
		fmt.Printf("Debug: API Error: %+v\n", err)
		fmt.Println("Error:", err)
		return
	}
	if len(result.Nodes) == 0 {
		fmt.Println("No nodes found matching the criteria")
		return
	}

	fmt.Printf("Found %d nodes:\n", len(result.Nodes))
	for _, node := range result.Nodes {
		title := node.Title
		if len(title) > 40 {
			title = fmt.Sprintf("%s...", title[:37])
		}
		fmt.Printf(" %-40s  %-12s  %5s  %s\n", title, node.UUID, node.HumanReadableSize(), node.Mimetype)
	}
}

func (c *FindCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&FindCommand{})
}

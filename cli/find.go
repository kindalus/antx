package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
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
	var filters antbox.NodeFilters

	if !strings.Contains(searchText, ",") {
		// Simple search: use title field for text search
		filters = extractSingleFilter(searchText)
	} else {
		// Complex search: parse comma-separated criteria
		var filterList antbox.NodeFilters1D
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
				filter := antbox.NodeFilter{tokens[0], antbox.FilterOperator(tokens[1]), value}
				filterList = append(filterList, filter)
			} else {
				// More than 3 tokens: field operator "rest as value"
				valueStr := strings.Join(tokens[2:], " ")
				value := convertValue(valueStr)
				filter := antbox.NodeFilter{tokens[0], antbox.FilterOperator(tokens[1]), value}
				filterList = append(filterList, filter)
			}
		}

		if len(filterList) == 0 {
			fmt.Println("No valid filters found in criteria")
			return
		}

		filters = filterList
	}

	result, err := client.FindNodes(filters, 20, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(result.Nodes) == 0 {
		fmt.Println("No nodes found matching the criteria")
		return
	}

	fmt.Printf("Found %d nodes:\n", len(result.Nodes))
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "UUID", "SIZE", "MODIFIED", "MIMETYPE", "TITLE")
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "----", "----", "--------", "--------", "-----")

	// Sort nodes: directories first, then files, both alphabetically by title
	sortedNodes := sortNodesForListing(result.Nodes)

	for _, node := range sortedNodes {
		// Format UUID (first 12 characters)
		uuid := node.UUID
		if len(uuid) > 12 {
			uuid = uuid[:12]
		}

		// Format size with padding
		size := node.HumanReadableSize()

		// Format modified date
		modifiedAt := formatModifiedDate(node.ModifiedAt)

		// Format mimetype (max 30 characters with ellipsis)
		mimetype := node.Mimetype
		if len(mimetype) > 30 {
			mimetype = fmt.Sprintf("%s...", mimetype[:27])
		}

		// Title is free form (no truncation)
		title := node.Title

		fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", uuid, size, modifiedAt, mimetype, title)
	}
}

func (c *FindCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&FindCommand{})
}

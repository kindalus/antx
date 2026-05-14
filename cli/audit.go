package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/c-bata/go-prompt"
)

type AuditCommand struct{}

func (c *AuditCommand) GetName() string {
	return "audit"
}

func (c *AuditCommand) GetDescription() string {
	return "Show node audit history"
}

func (c *AuditCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: audit [-v] <uuid>")
		return
	}

	verbose := false
	var nodeUUID string
	for _, arg := range args {
		switch arg {
		case "-v", "--verbose":
			verbose = true
		default:
			if nodeUUID == "" {
				nodeUUID = arg
			}
		}
	}
	if nodeUUID == "" {
		fmt.Println("Usage: audit [-v] <uuid>")
		return
	}
	node, err := client.GetNode(nodeUUID)
	if err != nil {
		fmt.Println("Error getting node:", err)
		return
	}
	if node.Mimetype == "" {
		fmt.Println("Error: node mimetype is required for audit lookup")
		return
	}

	events, err := client.GetAuditLog(nodeUUID, node.Mimetype)
	if err != nil {
		fmt.Println("Error getting audit log:", err)
		return
	}

	fmt.Printf("Audit log for %s (%s, %s)\n\n", node.Title, node.UUID, node.Mimetype)
	if len(events) == 0 {
		fmt.Println("No audit events found.")
		return
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Sequence > events[j].Sequence
	})

	fmt.Printf("%-6s %-24s  %-24s  %s\n", "SEQ", "EVENT", "OCCURRED", "USER")
	fmt.Printf("%-6s %-24s  %-24s  %s\n", "-----", "-----", "--------", "----")

	for _, event := range events {
		fmt.Printf("#%-5d %s  %s  %s\n", event.Sequence, event.EventType, event.OccurredOn, event.UserEmail)
		if verbose && len(event.Payload) > 0 {
			payload, err := json.MarshalIndent(event.Payload, "", "  ")
			if err != nil {
				fmt.Printf("Payload : %v\n", event.Payload)
			} else {
				fmt.Printf("Payload:\n%s\n", payload)
			}
		}
		if verbose {
			fmt.Println()
		}
	}
}

func (c *AuditCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&AuditCommand{})
}

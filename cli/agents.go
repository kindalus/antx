package cli

import (
	"fmt"
	"sort"

	"github.com/c-bata/go-prompt"
)

type AgentsCommand struct{}

func (c *AgentsCommand) GetName() string {
	return "agents"
}

func (c *AgentsCommand) GetDescription() string {
	return "List all available agents"
}

func (c *AgentsCommand) Execute(args []string) {
	agents, err := client.ListAgents()
	if err != nil {
		fmt.Println("Error listing agents:", err)
		return
	}

	if len(agents) == 0 {
		fmt.Println("No agents available.")
		return
	}

	// Sort agents alphabetically by title
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Title < agents[j].Title
	})

	fmt.Printf("Available agents (%d):\n", len(agents))
	fmt.Println()

	for _, agent := range agents {
		fmt.Printf("UUID: %s\n", agent.UUID)
		fmt.Printf("  Title: %s\n", agent.Title)
		if agent.Description != "" {
			fmt.Printf("  Description: %s\n", agent.Description)
		}
		if agent.Temperature > 0 {
			fmt.Printf("  Temperature: %.2f\n", agent.Temperature)
		}
		if agent.MaxTokens > 0 {
			fmt.Printf("  Max Tokens: %d\n", agent.MaxTokens)
		}
		fmt.Println()
	}
}

func (c *AgentsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	// This command doesn't take arguments, so no suggestions needed
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&AgentsCommand{})
}

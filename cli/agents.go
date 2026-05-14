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

	// Sort agents alphabetically by display name
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].DisplayName() < agents[j].DisplayName()
	})

	fmt.Printf("Available agents (%d):\n", len(agents))
	fmt.Println()

	for _, agent := range agents {
		fmt.Printf("UUID: %s\n", agent.UUID)
		fmt.Printf("  Name: %s\n", agent.DisplayName())
		if agent.Description != "" {
			fmt.Printf("  Description: %s\n", agent.Description)
		}
		if agent.Model != "" {
			fmt.Printf("  Model: %s\n", agent.Model)
		}
		if agent.MaxLlmCalls > 0 {
			fmt.Printf("  Max LLM Calls: %d\n", agent.MaxLlmCalls)
		}
		fmt.Printf("  Exposed To Users: %v\n", agent.ExposedToUsers)
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

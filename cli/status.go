package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type StatusCommand struct{}

func (c *StatusCommand) GetName() string {
	return "status"
}

func (c *StatusCommand) GetDescription() string {
	return "Show cached data statistics"
}

func (c *StatusCommand) Execute(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: status")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Display statistics about cached resources loaded at startup.")
		fmt.Println("  Shows the number of aspects, actions, extensions, and agents")
		fmt.Println("  currently available for auto-completion suggestions.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  status")
		return
	}

	aspects := GetCachedAspects()
	actions := GetCachedActions()
	extensions := GetCachedExtensions()
	agents := GetCachedAgents()

	fmt.Println("Cached Resource Statistics:")
	fmt.Println("========================================")
	fmt.Printf("  Aspects:    %d\n", len(aspects))
	fmt.Printf("  Actions:    %d\n", len(actions))
	fmt.Printf("  Extensions: %d\n", len(extensions))
	fmt.Printf("  Agents:     %d\n", len(agents))
	fmt.Println()

	total := len(aspects) + len(actions) + len(extensions) + len(agents)
	fmt.Printf("Total resources: %d\n", total)

	// Show conversation session statistics
	sessionCount := GetActiveSessionCount()
	fmt.Println()
	fmt.Println("Conversation Sessions:")
	fmt.Println("========================================")
	fmt.Printf("  Active sessions: %d\n", sessionCount)

	if sessionCount > 0 {
		sessions := ListActiveSessions()
		totalMessages := 0
		for _, sessionID := range sessions {
			session := GetOrCreateSession(sessionID)
			totalMessages += len(session.GetHistory())
		}
		fmt.Printf("  Total messages:  %d\n", totalMessages)

		fmt.Println()
		fmt.Println("  Recent sessions:")
		displayCount := sessionCount
		if displayCount > 5 {
			displayCount = 5
		}
		for i := 0; i < displayCount; i++ {
			sessionID := sessions[i]
			session := GetOrCreateSession(sessionID)
			messageCount := len(session.GetHistory())
			fmt.Printf("    %s (%d messages)\n", sessionID, messageCount)
		}
		if sessionCount > 5 {
			fmt.Printf("    ... and %d more (use 'sessions list' to see all)\n", sessionCount-5)
		}
	} else {
		fmt.Println("  No active conversation sessions")
		fmt.Println("  Start a conversation using 'chat' or 'rag' with -c <session_id>")
	}

	if total == 0 {
		fmt.Println()
		fmt.Println("Note: No resources loaded. This might indicate:")
		fmt.Println("  - Connection issues during startup")
		fmt.Println("  - No resources available on the server")
		fmt.Println("  - Authentication problems")
		fmt.Println()
		fmt.Println("Try running 'reload' to refresh the cache.")
	}
}

func (c *StatusCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&StatusCommand{})
}

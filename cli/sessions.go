package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type SessionsCommand struct{}

func (c *SessionsCommand) GetName() string {
	return "sessions"
}

func (c *SessionsCommand) GetDescription() string {
	return "Manage conversation sessions"
}

func (c *SessionsCommand) Execute(args []string) {
	if len(args) == 0 {
		c.showUsage()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		c.listSessions()
	case "clear":
		if len(args) < 2 {
			fmt.Println("Usage: sessions clear <session_id>")
			fmt.Println("       sessions clear all")
			return
		}
		sessionID := args[1]
		if sessionID == "all" {
			c.clearAllSessions()
		} else {
			c.clearSession(sessionID)
		}
	case "show":
		if len(args) < 2 {
			fmt.Println("Usage: sessions show <session_id>")
			return
		}
		c.showSession(args[1])
	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: sessions remove <session_id>")
			fmt.Println("       sessions remove all")
			return
		}
		sessionID := args[1]
		if sessionID == "all" {
			c.removeAllSessions()
		} else {
			c.removeSession(sessionID)
		}
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		c.showUsage()
	}
}

func (c *SessionsCommand) showUsage() {
	fmt.Println("Usage: sessions <subcommand> [args]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list                    List all active sessions")
	fmt.Println("  show <session_id>       Show conversation history for a session")
	fmt.Println("  clear <session_id>      Clear history for a session (keeps session active)")
	fmt.Println("  clear all               Clear history for all sessions")
	fmt.Println("  remove <session_id>     Remove a session completely")
	fmt.Println("  remove all              Remove all sessions")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  sessions list")
	fmt.Println("  sessions show my-chat")
	fmt.Println("  sessions clear my-chat")
	fmt.Println("  sessions remove old-session")
}

func (c *SessionsCommand) listSessions() {
	sessions := ListActiveSessions()
	count := GetActiveSessionCount()

	if count == 0 {
		fmt.Println("No active sessions.")
		return
	}

	fmt.Printf("Active sessions (%d):\n", count)
	for _, sessionID := range sessions {
		session := GetOrCreateSession(sessionID)
		historyCount := len(session.GetHistory())
		fmt.Printf("  %s (%d messages)\n", sessionID, historyCount)
	}
}

func (c *SessionsCommand) showSession(sessionID string) {
	session := GetOrCreateSession(sessionID)
	history := session.GetHistory()

	if len(history) == 0 {
		fmt.Printf("Session '%s' has no conversation history.\n", sessionID)
		return
	}

	fmt.Printf("Conversation history for session '%s':\n", sessionID)
	fmt.Println(strings.Repeat("-", 50))

	for i, msg := range history {
		fmt.Printf("[%d] %s: ", i+1, strings.ToUpper(msg.Role))

		switch content := msg.Content.(type) {
		case string:
			// Truncate long messages for display
			if len(content) > 200 {
				fmt.Printf("%s...\n", content[:200])
			} else {
				fmt.Println(content)
			}
		default:
			fmt.Printf("%v\n", content)
		}

		if i < len(history)-1 {
			fmt.Println()
		}
	}
}

func (c *SessionsCommand) clearSession(sessionID string) {
	// Check if session exists and has content
	session := GetOrCreateSession(sessionID)
	if session.IsEmpty() {
		fmt.Printf("Session '%s' is already empty.\n", sessionID)
		return
	}

	ClearSession(sessionID)
	fmt.Printf("Cleared conversation history for session '%s'.\n", sessionID)
}

func (c *SessionsCommand) clearAllSessions() {
	sessions := ListActiveSessions()
	if len(sessions) == 0 {
		fmt.Println("No active sessions to clear.")
		return
	}

	clearedCount := 0
	for _, sessionID := range sessions {
		session := GetOrCreateSession(sessionID)
		if !session.IsEmpty() {
			ClearSession(sessionID)
			clearedCount++
		}
	}

	fmt.Printf("Cleared conversation history for %d session(s).\n", clearedCount)
}

func (c *SessionsCommand) removeSession(sessionID string) {
	sessions := ListActiveSessions()
	found := false
	for _, id := range sessions {
		if id == sessionID {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Session '%s' not found.\n", sessionID)
		return
	}

	RemoveSession(sessionID)
	fmt.Printf("Removed session '%s'.\n", sessionID)
}

func (c *SessionsCommand) removeAllSessions() {
	sessions := ListActiveSessions()
	if len(sessions) == 0 {
		fmt.Println("No active sessions to remove.")
		return
	}

	count := len(sessions)
	for _, sessionID := range sessions {
		RemoveSession(sessionID)
	}

	fmt.Printf("Removed %d session(s).\n", count)
}

func (c *SessionsCommand) Suggest(d prompt.Document) []prompt.Suggest {
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
		// Suggesting subcommands
		currentWord := d.GetWordBeforeCursor()
		subcommands := []prompt.Suggest{
			{Text: "list", Description: "List all active sessions"},
			{Text: "show", Description: "Show conversation history for a session"},
			{Text: "clear", Description: "Clear history for a session"},
			{Text: "remove", Description: "Remove a session completely"},
		}

		var filtered []prompt.Suggest
		for _, cmd := range subcommands {
			if strings.HasPrefix(strings.ToLower(cmd.Text), strings.ToLower(currentWord)) {
				filtered = append(filtered, cmd)
			}
		}
		return filtered

	case 1:
		// Suggesting session IDs or 'all' for applicable commands
		if len(args) >= 2 {
			subcommand := args[1]
			if subcommand == "show" || subcommand == "clear" || subcommand == "remove" {
				currentWord := d.GetWordBeforeCursor()
				var suggests []prompt.Suggest

				// Add 'all' option for clear and remove commands
				if subcommand == "clear" || subcommand == "remove" {
					if strings.HasPrefix("all", strings.ToLower(currentWord)) {
						suggests = append(suggests, prompt.Suggest{
							Text:        "all",
							Description: "All sessions",
						})
					}
				}

				// Add active session IDs
				sessions := ListActiveSessions()
				for _, sessionID := range sessions {
					if strings.HasPrefix(strings.ToLower(sessionID), strings.ToLower(currentWord)) {
						session := GetOrCreateSession(sessionID)
						historyCount := len(session.GetHistory())
						suggests = append(suggests, prompt.Suggest{
							Text:        sessionID,
							Description: fmt.Sprintf("%d messages", historyCount),
						})
					}
				}
				return suggests
			}
		}
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&SessionsCommand{})
}

package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type agentChatRunner struct{}

// startInteractiveSession starts an interactive chat session with the specified agent
func (c *agentChatRunner) startInteractiveSession(agentUUID string, initialMessage string, temperature *float64, maxTokens *int) {
	// Find agent name for display
	agentName := agentUUID
	for _, agent := range GetCachedAgents() {
		if agent.UUID == agentUUID {
			agentName = agent.DisplayName()
			break
		}
	}

	fmt.Printf("Starting interactive chat with %s\n", agentName)
	fmt.Println("Type 'exit' or press Ctrl+D to exit the session.")
	fmt.Println()

	// Send initial message if provided
	if initialMessage != "" {
		fmt.Printf("You: %s\n", initialMessage)
		c.sendMessage(agentUUID, initialMessage, temperature, maxTokens)
		fmt.Println()
	}

	// Create interactive session context
	sessionContext := &ChatSessionContext{
		agentUUID:   agentUUID,
		temperature: temperature,
		maxTokens:   maxTokens,
		command:     c,
	}

	// Create a new prompt for the chat session
	p := prompt.New(
		sessionContext.executeMessage,
		func(d prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
		prompt.OptionTitle(fmt.Sprintf("Chat with %s", agentName)),
		prompt.OptionPrefix("You: "),
	)
	p.Run()
}

// ChatSessionContext holds the context for an interactive chat session
type ChatSessionContext struct {
	agentUUID   string
	temperature *float64
	maxTokens   *int
	command     *agentChatRunner
}

func (ctx *ChatSessionContext) executeMessage(input string) {
	input = strings.TrimSpace(input)

	// Check for exit command
	if input == "exit" {
		fmt.Println("Exiting chat session...")
		return
	}

	// Skip empty messages
	if input == "" {
		return
	}

	// Send message and display response
	ctx.command.sendMessage(ctx.agentUUID, input, ctx.temperature, ctx.maxTokens)
}

// sendMessage sends a single message to the agent and displays the response
func (c *agentChatRunner) sendMessage(agentUUID string, message string, temperature *float64, maxTokens *int) {
	// Find agent name for display
	agentName := agentUUID
	for _, agent := range GetCachedAgents() {
		if agent.UUID == agentUUID {
			agentName = agent.DisplayName()
			break
		}
	}

	// Show loading animation while waiting for response (dots style is less distracting in chat)
	animation := StartLoadingAnimationWithStyle(fmt.Sprintf("Chatting with %s", agentName), DotsStyle)
	chatHistory, err := client.ChatWithAgent(agentUUID, message, "", temperature, maxTokens, nil)

	if err != nil {
		animation.StopWithMessage(fmt.Sprintf("✗ Error chatting with %s", agentName))
		fmt.Println("Error:", err)
		return
	}

	animation.StopWithMessage(fmt.Sprintf("✓ %s:", agentName))

	// Find the last model response from the chat history
	for i := len(chatHistory) - 1; i >= 0; i-- {
		msg := chatHistory[i]
		if msg.Role == "model" {
			for _, part := range msg.Parts {
				if part.Text != nil {
					fmt.Printf("Assistant: %s\n", *part.Text)
					return
				}
			}
		}
	}

	// If no model response found, show that no response was received
	fmt.Println("Assistant: (no response)")
}

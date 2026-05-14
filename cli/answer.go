package cli

import (
	"fmt"
)

type agentAnswerRunner struct{}

func (c *agentAnswerRunner) askAgent(agentUUID string, question string, temperature *float64, maxTokens *int) {
	// Find agent name for display
	agentName := agentUUID
	for _, agent := range GetCachedAgents() {
		if agent.UUID == agentUUID {
			agentName = agent.DisplayName()
			break
		}
	}

	// Show loading animation while waiting for response
	animation := StartLoadingAnimationWithStyle(fmt.Sprintf("Asking %s", agentName), SpinnerStyle)
	chatHistory, err := client.AnswerFromAgent(agentUUID, question, temperature, maxTokens)

	if err != nil {
		animation.StopWithMessage(fmt.Sprintf("✗ Error asking %s", agentName))
		fmt.Println("Error:", err)
		return
	}

	animation.StopWithMessage(fmt.Sprintf("✓ Response from %s:", agentName))

	// Find the last model response from the chat history
	for i := len(chatHistory) - 1; i >= 0; i-- {
		msg := chatHistory[i]
		if msg.Role == "model" {
			for _, part := range msg.Parts {
				if part.Text != nil {
					fmt.Println(*part.Text)
					return
				}
			}
		}
	}

	// If no model response found, show that no response was received
	fmt.Println("(no response)")
}

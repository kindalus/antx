package cli

import (
	"testing"
)

func TestSessionManager(t *testing.T) {
	// Create a new session manager for testing
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}

	// Test getting a new session
	session1 := sm.GetSession("test-session-1")
	if session1 == nil {
		t.Fatal("Expected session to be created, got nil")
	}
	if session1.ID != "test-session-1" {
		t.Errorf("Expected session ID 'test-session-1', got '%s'", session1.ID)
	}
	if !session1.IsEmpty() {
		t.Error("Expected new session to be empty")
	}

	// Test getting the same session returns the same instance
	session1Again := sm.GetSession("test-session-1")
	if session1 != session1Again {
		t.Error("Expected to get the same session instance")
	}

	// Test session count
	if sm.GetSessionCount() != 1 {
		t.Errorf("Expected session count to be 1, got %d", sm.GetSessionCount())
	}
}

func TestSessionHistory(t *testing.T) {
	session := &Session{
		ID:      "test",
		History: []ConversationHistory{},
	}

	// Test empty session
	if !session.IsEmpty() {
		t.Error("Expected empty session")
	}
	history := session.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(history))
	}

	// Test adding messages
	session.AddMessage("user", "Hello")
	if session.IsEmpty() {
		t.Error("Expected session to not be empty after adding message")
	}

	session.AddMessage("assistant", "Hi there!")
	history = session.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	// Verify message content
	if history[0].Role != "user" || history[0].Content != "Hello" {
		t.Error("First message content incorrect")
	}
	if history[1].Role != "assistant" || history[1].Content != "Hi there!" {
		t.Error("Second message content incorrect")
	}

	// Test GetHistoryAsMap
	mapHistory := session.GetHistoryAsMap()
	if len(mapHistory) != 2 {
		t.Errorf("Expected 2 messages in map format, got %d", len(mapHistory))
	}
	if mapHistory[0]["role"] != "user" || mapHistory[0]["content"] != "Hello" {
		t.Error("First message in map format incorrect")
	}
}

func TestSessionClear(t *testing.T) {
	session := &Session{
		ID:      "test",
		History: []ConversationHistory{},
	}

	// Add some messages
	session.AddMessage("user", "Message 1")
	session.AddMessage("assistant", "Response 1")

	if session.IsEmpty() {
		t.Error("Expected session to have messages")
	}

	// Clear session
	session.Clear()
	if !session.IsEmpty() {
		t.Error("Expected session to be empty after clear")
	}

	history := session.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d messages", len(history))
	}
}

func TestSessionManagerOperations(t *testing.T) {
	// Reset session manager for testing
	originalManager := sessionManager
	sessionManager = &SessionManager{
		sessions: make(map[string]*Session),
	}
	defer func() {
		sessionManager = originalManager
	}()

	// Test global functions
	session := GetOrCreateSession("global-test")
	if session == nil {
		t.Fatal("Expected session to be created")
	}

	// Test adding messages
	AddMessageToSession("global-test", "user", "Test message")
	AddMessageToSession("global-test", "assistant", "Test response")

	history := GetSessionHistory("global-test")
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	// Test session count
	if GetActiveSessionCount() != 1 {
		t.Errorf("Expected 1 active session, got %d", GetActiveSessionCount())
	}

	// Test list sessions
	sessions := ListActiveSessions()
	if len(sessions) != 1 || sessions[0] != "global-test" {
		t.Errorf("Expected ['global-test'], got %v", sessions)
	}

	// Test clear session
	if IsSessionEmpty("global-test") {
		t.Error("Expected session to not be empty")
	}
	ClearSession("global-test")
	if !IsSessionEmpty("global-test") {
		t.Error("Expected session to be empty after clear")
	}

	// Session should still exist after clear
	if GetActiveSessionCount() != 1 {
		t.Errorf("Expected 1 active session after clear, got %d", GetActiveSessionCount())
	}

	// Test remove session
	RemoveSession("global-test")
	if GetActiveSessionCount() != 0 {
		t.Errorf("Expected 0 active sessions after remove, got %d", GetActiveSessionCount())
	}
}

func TestConcurrentAccess(t *testing.T) {
	session := &Session{
		ID:      "concurrent-test",
		History: []ConversationHistory{},
	}

	// Test concurrent writes (basic smoke test)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			session.AddMessage("user", "Message from goroutine 1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			session.AddMessage("assistant", "Response from goroutine 2")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should have 20 messages total
	history := session.GetHistory()
	if len(history) != 20 {
		t.Errorf("Expected 20 messages after concurrent writes, got %d", len(history))
	}
}

func TestSessionHistoryIntegrity(t *testing.T) {
	session := &Session{
		ID:      "integrity-test",
		History: []ConversationHistory{},
	}

	// Add messages in a specific pattern
	messages := []struct {
		role    string
		content string
	}{
		{"user", "First user message"},
		{"assistant", "First assistant response"},
		{"user", "Second user message"},
		{"assistant", "Second assistant response"},
	}

	for _, msg := range messages {
		session.AddMessage(msg.role, msg.content)
	}

	// Get history and verify order and content
	history := session.GetHistory()
	if len(history) != len(messages) {
		t.Fatalf("Expected %d messages, got %d", len(messages), len(history))
	}

	for i, msg := range messages {
		if history[i].Role != msg.role {
			t.Errorf("Message %d: expected role '%s', got '%s'", i, msg.role, history[i].Role)
		}
		if history[i].Content != msg.content {
			t.Errorf("Message %d: expected content '%s', got '%s'", i, msg.content, history[i].Content)
		}
	}

	// Test that GetHistory returns a copy (modifications don't affect original)
	historyOriginal := session.GetHistory()
	historyCopy := session.GetHistory()

	// Modify the copy
	if len(historyCopy) > 0 {
		historyCopy[0].Role = "modified"
	}

	// Original should be unchanged
	historyAfterModification := session.GetHistory()
	if len(historyAfterModification) > 0 && historyAfterModification[0].Role == "modified" {
		t.Error("GetHistory should return a copy, but original was modified")
	}

	// Verify original is still correct
	if historyOriginal[0].Role != messages[0].role {
		t.Error("Original history was unexpectedly modified")
	}
}

func TestComplexContentTypes(t *testing.T) {
	session := &Session{
		ID:      "complex-content-test",
		History: []ConversationHistory{},
	}

	// Test different content types
	session.AddMessage("user", "Simple string")
	session.AddMessage("assistant", map[string]any{
		"response": "Complex response",
		"metadata": map[string]string{
			"source": "test",
		},
	})
	session.AddMessage("user", []any{"array", "content", 123})

	history := session.GetHistoryAsMap()
	if len(history) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(history))
	}

	// Verify content types are preserved
	if history[0]["content"] != "Simple string" {
		t.Error("String content not preserved")
	}

	complexContent, ok := history[1]["content"].(map[string]any)
	if !ok {
		t.Error("Complex content type not preserved")
	} else if complexContent["response"] != "Complex response" {
		t.Error("Complex content value not preserved")
	}

	arrayContent, ok := history[2]["content"].([]any)
	if !ok {
		t.Error("Array content type not preserved")
	} else if len(arrayContent) != 3 {
		t.Error("Array content length not preserved")
	}
}

package cli

import (
	"sync"
)

// ConversationHistory represents a single message in the conversation
type ConversationHistory struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content any    `json:"content"` // message content
}

// Session represents a conversation session with history
type Session struct {
	ID      string                `json:"id"`
	History []ConversationHistory `json:"history"`
	mu      sync.RWMutex          `json:"-"`
}

// SessionManager manages all active sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var (
	sessionManager = &SessionManager{
		sessions: make(map[string]*Session),
	}
)

// GetSession retrieves or creates a session by ID
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if exists {
		return session
	}

	// Create new session if it doesn't exist
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Double-check after acquiring write lock
	if session, exists := sm.sessions[sessionID]; exists {
		return session
	}

	session = &Session{
		ID:      sessionID,
		History: []ConversationHistory{},
	}
	sm.sessions[sessionID] = session
	return session
}

// AddMessage adds a message to the session history
func (s *Session) AddMessage(role string, content any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.History = append(s.History, ConversationHistory{
		Role:    role,
		Content: content,
	})
}

// GetHistory returns a copy of the conversation history
func (s *Session) GetHistory() []ConversationHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]ConversationHistory, len(s.History))
	copy(history, s.History)
	return history
}

// GetHistoryAsMap returns the history in the format expected by the API
func (s *Session) GetHistoryAsMap() []map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var history []map[string]any
	for _, msg := range s.History {
		history = append(history, map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	return history
}

// IsEmpty checks if the session has no history
func (s *Session) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.History) == 0
}

// Clear removes all history from the session
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = []ConversationHistory{}
}

// RemoveSession removes a session from the manager
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// ListSessions returns all active session IDs
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []string
	for id := range sm.sessions {
		sessions = append(sessions, id)
	}
	return sessions
}

// GetSessionCount returns the number of active sessions
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Global session management functions

// GetOrCreateSession gets or creates a session for the given conversation ID
func GetOrCreateSession(conversationID string) *Session {
	return sessionManager.GetSession(conversationID)
}

// AddMessageToSession adds a message to the specified session
func AddMessageToSession(conversationID, role string, content any) {
	session := sessionManager.GetSession(conversationID)
	session.AddMessage(role, content)
}

// GetSessionHistory returns the history for the specified session
func GetSessionHistory(conversationID string) []map[string]any {
	session := sessionManager.GetSession(conversationID)
	return session.GetHistoryAsMap()
}

// IsSessionEmpty checks if the specified session is empty
func IsSessionEmpty(conversationID string) bool {
	session := sessionManager.GetSession(conversationID)
	return session.IsEmpty()
}

// ClearSession clears the history for the specified session
func ClearSession(conversationID string) {
	session := sessionManager.GetSession(conversationID)
	session.Clear()
}

// RemoveSession removes the specified session
func RemoveSession(conversationID string) {
	sessionManager.RemoveSession(conversationID)
}

// GetActiveSessionCount returns the number of active sessions
func GetActiveSessionCount() int {
	return sessionManager.GetSessionCount()
}

// ListActiveSessions returns all active session IDs
func ListActiveSessions() []string {
	return sessionManager.ListSessions()
}

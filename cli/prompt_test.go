package cli

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/kindalus/antx/antbox"

	prompt "github.com/c-bata/go-prompt"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writer
	fn()
	writer.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	return string(output)
}

type mockClient struct {
}

func (c *mockClient) Login() error {
	return nil
}

func (c *mockClient) GetCurrentUser() (*antbox.User, error) {
	return &antbox.User{
		Email:  "test@example.com",
		Name:   "Test User",
		Groups: []string{"testers"},
	}, nil
}

func (c *mockClient) GetNode(uuid string) (*antbox.Node, error) {
	if uuid == "--root--" {
		return &antbox.Node{UUID: "--root--", Title: "root", Parent: "", Mimetype: "application/vnd.antbox.folder"}, nil
	}
	return &antbox.Node{UUID: "test-uuid", Title: "test-title", Parent: "--root--", Mimetype: "application/pdf"}, nil
}

func (c *mockClient) ListNodes(parent string) ([]antbox.Node, error) {
	return []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}, nil
}

func (c *mockClient) CreateFolder(parent, name string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.folder"}, nil
}

func (c *mockClient) CreateSmartFolder(parent, name string, filters antbox.NodeFilters) (*antbox.Node, error) {
	return &antbox.Node{UUID: "smart-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.smartfolder"}, nil
}

func (c *mockClient) SetAuthHeader(req *http.Request) {}

func (c *mockClient) RemoveNode(uuid string) error {
	return nil
}

func (c *mockClient) MoveNode(uuid, newParent string) error {
	return nil
}

func (c *mockClient) ChangeNodeName(uuid, newName string) error {
	return nil
}

func (c *mockClient) CreateFile(filePath string, metadata antbox.NodeCreate) (*antbox.Node, error) {
	parentUuid := metadata.Parent
	return &antbox.Node{UUID: "uploaded-uuid", Title: "uploaded-file.txt", Parent: parentUuid}, nil
}

func (c *mockClient) CreateNode(node antbox.NodeCreate) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-node-uuid", Title: node.Title, Parent: node.Parent, Mimetype: node.Mimetype}, nil
}

func (c *mockClient) UpdateNode(uuid string, metadata antbox.NodeUpdate) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-title"}, nil
}

func (c *mockClient) UpdateFile(uuid, filePath string) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-file.txt", Parent: "--root--"}, nil
}

func (c *mockClient) FindNodes(filters antbox.NodeFilters, pageSize, pageToken int) (*antbox.NodeFilterResult, error) {
	return &antbox.NodeFilterResult{
		Nodes:     []antbox.Node{{UUID: "found-uuid", Title: "found-node", Mimetype: "text/plain"}},
		PageSize:  pageSize,
		PageToken: pageToken,
	}, nil
}

func (c *mockClient) EvaluateNode(uuid string) ([]antbox.Node, error) {
	// Return mock nodes for smartfolder evaluation
	return []antbox.Node{
		{UUID: "eval-node-1", Title: "Evaluated Node 1", Mimetype: "text/plain"},
		{UUID: "eval-node-2", Title: "Evaluated Node 2", Mimetype: "application/pdf"},
	}, nil
}

func (c *mockClient) DownloadNode(uuid, downloadPath string) error {
	return nil
}

func (c *mockClient) GetBreadcrumbs(uuid string) ([]antbox.Node, error) {
	return []antbox.Node{
		{UUID: "--root--", Title: "root", Parent: ""},
		{UUID: "test-uuid", Title: "test-title", Parent: "--root--"},
	}, nil
}

func (c *mockClient) ChatWithAgent(agentUUID string, message string, conversationID string, temperature *float64, maxTokens *int, history []map[string]any) (antbox.ChatHistory, error) {
	text := "Mock chat response"
	return antbox.ChatHistory{
		{
			Role: antbox.ChatMessageRoleModel,
			Parts: []antbox.ChatMessagePart{
				{Text: &text},
			},
		},
	}, nil
}

func (c *mockClient) AnswerFromAgent(agentUUID string, query string, temperature *float64, maxTokens *int) (antbox.ChatHistory, error) {
	text := "Mock answer response"
	return antbox.ChatHistory{
		{
			Role: antbox.ChatMessageRoleModel,
			Parts: []antbox.ChatMessagePart{
				{Text: &text},
			},
		},
	}, nil
}

// New interface methods
func (c *mockClient) CopyNode(uuid, parent, title string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "copied-uuid", Title: title, Parent: parent}, nil
}

func (c *mockClient) CloneNode(uuid string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "cloned-uuid", Title: "Clone of test-title", Parent: "--root--"}, nil
}

func (c *mockClient) ExportNode(uuid string, format string) ([]byte, error) {
	return []byte("exported content"), nil
}

func (c *mockClient) ListAgents() ([]antbox.Agent, error) {
	return []antbox.Agent{{UUID: "agent-uuid", Name: "Test Agent"}}, nil
}

func (c *mockClient) UploadAgent(filePath string) (*antbox.Agent, error) {
	return &antbox.Agent{UUID: "agent-uuid", Name: "test-agent"}, nil
}

func (c *mockClient) GetAgent(uuid string) (*antbox.Agent, error) {
	return &antbox.Agent{UUID: uuid, Name: "Test Agent"}, nil
}

func (c *mockClient) DeleteAgent(uuid string) error {
	return nil
}

func (c *mockClient) ListAPIKeys() ([]antbox.APIKey, error) {
	return []antbox.APIKey{{UUID: "api-key-uuid", Description: "Test API Key"}}, nil
}

func (c *mockClient) CreateAPIKey(request antbox.APIKeyCreate) (*antbox.APIKey, error) {
	return &antbox.APIKey{UUID: "new-api-key-uuid", Description: request.Description}, nil
}

func (c *mockClient) GetAPIKey(uuid string) (*antbox.APIKey, error) {
	return &antbox.APIKey{UUID: uuid, Description: "Test API Key"}, nil
}

func (c *mockClient) DeleteAPIKey(uuid string) error {
	return nil
}

func (c *mockClient) ListUsers() ([]antbox.User, error) {
	return []antbox.User{{UUID: "user-uuid", Email: "test@example.com"}}, nil
}

func (c *mockClient) CreateUser(user antbox.UserCreate) (*antbox.User, error) {
	return &antbox.User{UUID: "new-user-uuid", Email: user.Email}, nil
}

func (c *mockClient) GetUser(email string) (*antbox.User, error) {
	return &antbox.User{UUID: "user-uuid", Email: email}, nil
}

func (c *mockClient) UpdateUser(email string, user antbox.UserUpdate) (*antbox.User, error) {
	return &antbox.User{UUID: "user-uuid", Email: email}, nil
}

func (c *mockClient) DeleteUser(uuid string) error {
	return nil
}

func (c *mockClient) ListGroups() ([]antbox.Group, error) {
	return []antbox.Group{{UUID: "group-uuid", Title: "Test Group"}}, nil
}

func (c *mockClient) CreateGroup(group antbox.GroupCreate) (*antbox.Group, error) {
	return &antbox.Group{UUID: "new-group-uuid", Title: group.Title}, nil
}

func (c *mockClient) GetGroup(uuid string) (*antbox.Group, error) {
	return &antbox.Group{UUID: uuid, Title: "Test Group"}, nil
}

func (c *mockClient) DeleteGroup(uuid string) error {
	return nil
}

func (c *mockClient) ListDocs() ([]antbox.DocInfo, error) {
	return []antbox.DocInfo{{UUID: "doc-uuid", Description: "Test Documentation"}}, nil
}

func (c *mockClient) GetDoc(uuid string) (string, error) {
	return "# Test Documentation\n\nThis is test documentation content.", nil
}

func (c *mockClient) GetAuditLog(uuid, mimetype string) ([]antbox.AuditEvent, error) {
	return []antbox.AuditEvent{
		{
			StreamID:   uuid,
			EventID:    "event-1",
			EventType:  "NodeUpdatedEvent",
			OccurredOn: "2026-05-14T10:00:00Z",
			UserEmail:  "user@example.com",
			Tenant:     "default",
			Payload:    map[string]any{"title": map[string]any{"old": "Old", "new": "New"}},
			Sequence:   1,
		},
		{
			StreamID:   uuid,
			EventID:    "event-2",
			EventType:  "NodeDeletedEvent",
			OccurredOn: "2026-05-14T11:00:00Z",
			UserEmail:  "admin@example.com",
			Tenant:     "default",
			Payload:    map[string]any{"deleted": true},
			Sequence:   2,
		},
	}, nil
}

func TestExecutor(t *testing.T) {
	client = &mockClient{}

	executor("ls")
	executor("pwd")
	executor("stat test-uuid")
	executor("cd test-uuid")
	executor("cd ..")
	executor("mkdir test-folder")
	executor("mksmart Facturas_2025 title match comprovativo")
	executor("update test-uuid /path/to/file.txt")
}

// Helper function to create a properly configured Document for testing
func createTestDocument(text string) prompt.Document {
	doc := prompt.Document{Text: text}

	// Use reflection to set the private cursorPosition field to the end of text
	v := reflect.ValueOf(&doc).Elem()
	cursorField := v.FieldByName("cursorPosition")

	if cursorField.IsValid() {
		// Make the field accessible
		cursorField = reflect.NewAt(cursorField.Type(), unsafe.Pointer(cursorField.UnsafeAddr())).Elem()
		cursorField.SetInt(int64(len(text)))
	}

	return doc
}

func TestCompleter(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"}}

	// Test with single character - should return command suggestions for prefix matches
	doc := createTestDocument("l")
	suggests := completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'l', got %d", len(suggests))
	}
	if len(suggests) > 0 && suggests[0].Text != "ls" {
		t.Errorf("Expected 'ls' suggestion for 'l', got '%s'", suggests[0].Text)
	}

	// Test with exact command match - should return no suggestions
	doc = createTestDocument("ls")
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for exact command 'ls', got %d", len(suggests))
	}

	// Test ls command with arguments - ls doesn't provide node suggestions
	doc = createTestDocument("ls te")
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for 'ls te', got %d", len(suggests))
	}

	// Test cd command - should use UUID for folder suggestions
	doc = createTestDocument("cd te")
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'cd te', got %d. Current nodes: %+v", len(suggests), currentNodes)
		for i, s := range suggests {
			t.Errorf("Suggestion %d: Text='%s', Description='%s'", i, s.Text, s.Description)
		}
	}
	if len(suggests) > 0 {
		if suggests[0].Text != "test-uuid" {
			t.Errorf("Expected cd suggestion to use UUID 'test-uuid', got '%s'", suggests[0].Text)
		}
		if !strings.Contains(suggests[0].Description, "test-title") {
			t.Errorf("Expected cd suggestion description to contain folder name, got '%s'", suggests[0].Description)
		}
	}

	// Test command with single char argument - should return no suggestions (word too short)
	doc = createTestDocument("ls t")
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for single char argument, got %d", len(suggests))
	}
}

func TestCompleterCdWithMixedNodeTypes(t *testing.T) {
	client = &mockClient{}
	// Mix of folder and file nodes
	currentNodes = []antbox.Node{
		{UUID: "folder-uuid-1", Title: "documents", Mimetype: "application/vnd.antbox.folder", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
		{UUID: "file-uuid-1", Title: "document.txt", Mimetype: "text/plain", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
		{UUID: "folder-uuid-2", Title: "downloads", Mimetype: "application/vnd.antbox.folder", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
	}

	// Test cd command with folder prefix - should only suggest folders and use UUIDs
	doc := createTestDocument("cd do")
	suggests := completer(doc)

	// Should get 2 suggestions: documents and downloads (both folders)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 folder suggestions for 'cd do', got %d", len(suggests))
	}

	// All suggestions should be UUIDs for folders
	for _, suggest := range suggests {
		if suggest.Text != "folder-uuid-1" && suggest.Text != "folder-uuid-2" {
			t.Errorf("Expected folder UUID in suggestion text, got '%s'", suggest.Text)
		}
		// Description should contain the folder name
		if suggest.Description == "" {
			t.Errorf("Expected non-empty description, got '%s'", suggest.Description)
		}
	}

	// Test ls command with same prefix - ls doesn't provide node suggestions
	doc = createTestDocument("ls do")
	suggests = completer(doc)

	// ls command doesn't provide node suggestions, should get 0
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for 'ls do', got %d", len(suggests))
	}

	// Test cd with exact folder name
	doc = createTestDocument("cd documents")
	suggests = completer(doc)

	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for exact folder name, got %d", len(suggests))
	}

	if suggests[0].Text != "folder-uuid-1" {
		t.Errorf("Expected folder UUID 'folder-uuid-1', got '%s'", suggests[0].Text)
	}
}

func TestCompleterNewCommands(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{
		{UUID: "folder-uuid", Title: "test-folder", Mimetype: "application/vnd.antbox.folder", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
		{UUID: "file-uuid", Title: "test-file.txt", Mimetype: "text/plain", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
	}

	// Test rm command suggestions (should suggest all nodes)
	doc := createTestDocument("rm te")
	suggests := completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'rm te', got %d", len(suggests))
	}

	// Test mv command first argument (should suggest all nodes)
	doc = createTestDocument("mv te")
	suggests = completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'mv te', got %d", len(suggests))
	}

	// Test mv command second argument (should only suggest folders with UUID)
	doc = createTestDocument("mv test-uuid te")
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for mv second argument, got %d", len(suggests))
	}
	if suggests[0].Text != "folder-uuid" {
		t.Errorf("Expected folder UUID for mv destination, got '%s'", suggests[0].Text)
	}

	// Test cp command first argument (should not provide suggestions)
	doc = createTestDocument("cp /path/to/file")
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for cp file path, got %d", len(suggests))
	}

	// Test cp command - cp is implemented, should get folder suggestions only
	doc = createTestDocument("cp /path/to/file te")
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for cp command (folders only), got %d", len(suggests))
	}

	// Test stat command suggestions (should suggest all nodes)
	doc = createTestDocument("stat te")
	suggests = completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'stat te', got %d", len(suggests))
	}

	// Test rename command suggestions (should suggest all nodes)
	doc = createTestDocument("rename te")
	suggests = completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'rename te', got %d", len(suggests))
	}
}

func TestCommandSuggestions(t *testing.T) {
	client = &mockClient{}

	// Test that all commands are suggested when typing partial matches
	testCases := []struct {
		input    string
		expected []string
	}{
		{"ls", []string{}}, // exact matches return no suggestions
		{"rm", []string{}}, // exact matches return no suggestions
		{"mk", []string{"mkdir", "mksmart"}},
		{"mv", []string{}}, // exact matches return no suggestions
		{"cd", []string{}}, // exact matches return no suggestions
		{"up", []string{"upload"}},
		{"ex", []string{"exit"}},
		{"he", []string{"help"}},
		{"pw", []string{"pwd"}},
		{"st", []string{"stat", "status"}},
		{"fi", []string{"find"}},
		{"re", []string{"reload", "rename"}},
		{"cl", []string{"clone"}},
		{"ch", []string{}},
		{"an", []string{}},
		{"ra", []string{}},
		{"do", []string{"docs", "download"}},
		{"te", []string{}},
		{"ag", []string{"agents"}},
		{"au", []string{"audit"}},
		{"ac", []string{}},
	}

	for _, tc := range testCases {
		doc := createTestDocument(tc.input)
		suggests := completer(doc)

		if len(suggests) != len(tc.expected) {
			t.Errorf("For input '%s': expected %d suggestions, got %d", tc.input, len(tc.expected), len(suggests))
			continue
		}

		// Check that all expected commands are present
		suggestedTexts := make([]string, len(suggests))
		for i, suggest := range suggests {
			suggestedTexts[i] = suggest.Text
		}

		for _, expectedCmd := range tc.expected {
			found := slices.Contains(suggestedTexts, expectedCmd)
			if !found && len(tc.expected) > 0 {
				t.Errorf("For input '%s': expected command '%s' not found in suggestions %v", tc.input, expectedCmd, suggestedTexts)
			}
		}
	}
}

func TestCommandSuggestionsMinLength(t *testing.T) {
	client = &mockClient{}

	// Test that single character inputs return no suggestions
	// Test inputs that should return command suggestions
	testInputs := []struct {
		input    string
		expected int
	}{
		{"l", 1}, // should match "ls"
		{"r", 3}, // should match "rm", "rename", "reload"
		{"m", 3}, // should match "mkdir", "mv", "mksmart"
		{"c", 3}, // should match "cd", "clone", "cp"
		{"e", 1}, // should match "exit"
		{"a", 3}, // should match "agents", "aliases", "audit"
		{"h", 2}, // should match "help", "history"
	}

	for _, test := range testInputs {
		doc := createTestDocument(test.input)
		suggests := completer(doc)

		if len(suggests) != test.expected {
			t.Errorf("For input '%s': expected %d suggestions, got %d", test.input, test.expected, len(suggests))
		}
	}

	// Test that exact command matches return no suggestions
	doc := createTestDocument("ls")
	suggests := completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for exact command 'ls', got %d suggestions", len(suggests))
	}
}

func TestUploadUpdateCommandSuggestions(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{
		{UUID: "folder-uuid", Title: "test-folder", Mimetype: "application/vnd.antbox.folder", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
		{UUID: "file-uuid", Title: "test-file.txt", Mimetype: "text/plain", CreatedAt: "2024-01-01T12:00:00Z", ModifiedAt: "2024-01-01T12:00:00Z"},
	}

	// Test upload -u command first argument (should suggest only files, not folders)
	doc := createTestDocument("upload -u te")
	suggests := completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'upload -u te', got %d", len(suggests))
	}

	// Test upload -u command suggestions should only include files
	foundFile := false
	foundFolder := false
	for _, suggest := range suggests {
		if suggest.Text == "file-uuid" {
			foundFile = true
		}
		if suggest.Text == "folder-uuid" {
			foundFolder = true
		}
	}

	if !foundFile {
		t.Errorf("Expected to find file-uuid in upload -u suggestions")
	}
	if foundFolder {
		t.Errorf("Did not expect to find folder-uuid in upload -u suggestions (should only suggest files)")
	}
}

func TestFindCommand(t *testing.T) {
	client = &mockClient{}

	// Test find with simple search
	doc := createTestDocument("find te")
	suggests := completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for find command, got %d", len(suggests))
	}
}

func TestUploadUpdateCommand(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{{UUID: "test-uuid", Title: "test-file.txt", Mimetype: "text/plain"}}

	// Test upload -u command first argument - should suggest nodes
	doc := createTestDocument("upload -u te")
	suggests := completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'upload -u te', got %d", len(suggests))
	}
	if len(suggests) > 0 && suggests[0].Text != "test-uuid" {
		t.Errorf("Expected suggestion 'test-uuid', got '%s'", suggests[0].Text)
	}
}

func TestUploadCommandFlags(t *testing.T) {
	client = &mockClient{}

	// Test upload command flag suggestions
	doc := createTestDocument("upload -")
	suggests := completer(doc)

	expectedFlags := []string{"-i", "-u"}
	if len(suggests) != len(expectedFlags) {
		t.Errorf("Expected %d flag suggestions, got %d", len(expectedFlags), len(suggests))
	}

	for i, expected := range expectedFlags {
		if i < len(suggests) && suggests[i].Text != expected {
			t.Errorf("Expected flag '%s' at position %d, got '%s'", expected, i, suggests[i].Text)
		}
	}
}

func TestCommandSuggestionsWithNewCommands(t *testing.T) {
	client = &mockClient{}

	testCases := []struct {
		input    string
		expected []string
	}{
		{"fi", []string{"find"}},
		{"up", []string{"upload"}},
	}

	for _, tc := range testCases {
		doc := createTestDocument(tc.input)
		suggests := completer(doc)

		if len(suggests) != len(tc.expected) {
			t.Errorf("For input '%s': expected %d suggestions, got %d", tc.input, len(tc.expected), len(suggests))
			continue
		}

		// Check that all expected commands are present
		suggestedTexts := make([]string, len(suggests))
		for i, suggest := range suggests {
			suggestedTexts[i] = suggest.Text
		}

		for _, expectedCmd := range tc.expected {
			found := slices.Contains(suggestedTexts, expectedCmd)
			if !found && len(tc.expected) > 0 {
				t.Errorf("For input '%s': expected command '%s' not found in suggestions %v", tc.input, expectedCmd, suggestedTexts)
			}
		}
	}
}

func TestRemovedCommandsAreNotRegistered(t *testing.T) {
	removedCommands := []string{"actions", "answer", "chat", "duplicate", "exec", "extensions", "rag", "run", "templates"}
	for _, name := range removedCommands {
		if _, exists := commands[name]; exists {
			t.Errorf("Expected %q command to be removed", name)
		}
	}

	if _, exists := commands["clone"]; !exists {
		t.Error("Expected clone command to remain registered")
	}
	if _, exists := commands["agents"]; !exists {
		t.Error("Expected agents command to remain registered")
	}
}

func TestAgentShortcutParsing(t *testing.T) {
	testCases := []struct {
		input     string
		mode      string
		agentUUID string
		message   string
		matched   bool
	}{
		{"/agent-uuid hello there", "chat", "agent-uuid", "hello there", true},
		{"/[agent-uuid] hello there", "chat", "agent-uuid", "hello there", true},
		{"@agent-uuid what is this?", "answer", "agent-uuid", "what is this?", true},
		{"@[agent-uuid] what is this?", "answer", "agent-uuid", "what is this?", true},
		{"ls", "", "", "", false},
	}

	for _, tc := range testCases {
		mode, agentUUID, message, matched := parseAgentShortcut(tc.input)
		if mode != tc.mode || agentUUID != tc.agentUUID || message != tc.message || matched != tc.matched {
			t.Errorf("parseAgentShortcut(%q) = (%q, %q, %q, %v), want (%q, %q, %q, %v)", tc.input, mode, agentUUID, message, matched, tc.mode, tc.agentUUID, tc.message, tc.matched)
		}
	}
}

func TestAgentShortcutSuggestions(t *testing.T) {
	cachedAgents = []antbox.Agent{
		{UUID: "agent-uuid", Name: "Helper Agent"},
		{UUID: "other-uuid", Name: "Other Agent"},
	}

	doc := createTestDocument("/agent")
	suggests := completer(doc)
	if len(suggests) != 1 || suggests[0].Text != "/[agent-uuid]" {
		t.Fatalf("Expected slash shortcut suggestion for agent-uuid, got %v", suggests)
	}

	doc = createTestDocument("@Helper")
	suggests = completer(doc)
	if len(suggests) != 1 || suggests[0].Text != "@[agent-uuid]" {
		t.Fatalf("Expected answer shortcut suggestion for agent-uuid, got %v", suggests)
	}
}

func TestListCommandsAfterRemoval(t *testing.T) {
	if agentsCmd, exists := commands["agents"]; exists {
		if agentsCmd.GetDescription() != "List all available agents" {
			t.Errorf("Unexpected agents description: %s", agentsCmd.GetDescription())
		}
	} else {
		t.Error("Expected agents command to be registered")
	}
}

func TestUpdatedCommandOutputs(t *testing.T) {
	client = &mockClient{}

	agentsCmd := &AgentsCommand{}
	if agentsCmd.GetDescription() != "List all available agents" {
		t.Error("Agents command description incorrect")
	}
}

func TestExtractSingleFilterUsesOpenAPIMatchOperator(t *testing.T) {
	filters := extractSingleFilter("title ~= document")
	if len(filters) != 1 {
		t.Fatalf("Expected one filter, got %#v", filters)
	}
	if filters[0][0] != "title" || filters[0][1] != antbox.FilterOperatorMatch || filters[0][2] != "document" {
		t.Fatalf("Expected ~= to normalize to OpenAPI match operator, got %#v", filters[0])
	}

	filters = extractSingleFilter("full text search")
	if len(filters) != 1 || filters[0][0] != ":content" || filters[0][1] != antbox.FilterOperatorMatch || filters[0][2] != "full text search" {
		t.Fatalf("Expected simple text search to become content match filter, got %#v", filters)
	}
}

func TestFindCommandParsing(t *testing.T) {
	client = &mockClient{}

	// Test the parsing functions directly
	testCases := []struct {
		input    string
		expected string
	}{
		{"title==Document", "title == Document"},
		{"size>1000", "size > 1000"},
		{"owner~=admin", "owner ~= admin"},
		{"createdTime>=2023-01-01", "createdTime >= 2023-01-01"},
		{"title == Document", "title == Document"}, // Already normalized
	}

	for _, tc := range testCases {
		result := normalizeOperators(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeOperators('%s') = '%s', expected '%s'", tc.input, result, tc.expected)
		}
	}

	// Test value conversion
	valueTests := []struct {
		input    string
		expected any
	}{
		{"123", 123},
		{"45.67", 45.67},
		{"true", true},
		{"false", false},
		{"hello", "hello"},
		{"2023-01-01", "2023-01-01"}, // Should remain string
	}

	for _, tc := range valueTests {
		result := convertValue(tc.input)
		if result != tc.expected {
			t.Errorf("convertValue('%s') = %v (%T), expected %v (%T)",
				tc.input, result, result, tc.expected, tc.expected)
		}
	}
}

func TestFindCommandCompleteWorkflow(t *testing.T) {
	client = &mockClient{}

	// Test complete parsing workflow for complex searches
	testCases := []struct {
		name          string
		args          []string
		expectedType  string
		expectedCount int
	}{
		{
			name:          "Simple search",
			args:          []string{"document"},
			expectedType:  "simple",
			expectedCount: 1,
		},
		{
			name:          "Complex search with multiple criteria",
			args:          []string{"title", "==", "Document,owner", "~=", "admin,size", ">", "1000"},
			expectedType:  "complex",
			expectedCount: 3,
		},
		{
			name:          "Complex search with spaces in values",
			args:          []string{"title", "==", "My", "Document,owner", "!=", "test", "user"},
			expectedType:  "complex",
			expectedCount: 2,
		},
		{
			name:          "Mixed operators without spaces",
			args:          []string{"title==Document,size>=1000,owner~=admin"},
			expectedType:  "complex",
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This would normally call the find function, but since we can't easily
			// test the full workflow without a real server, we test the parsing logic
			searchText := strings.Join(tc.args, " ")

			if !strings.Contains(searchText, ",") {
				// Simple search
				if tc.expectedType != "simple" {
					t.Errorf("Expected simple search for '%s'", searchText)
				}
			} else {
				// Complex search parsing
				if tc.expectedType != "complex" {
					t.Errorf("Expected complex search for '%s'", searchText)
				}

				var filterList [][]any
				parts := strings.Split(searchText, ",")

				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}

					part = normalizeOperators(part)
					tokens := strings.Fields(part)

					if len(tokens) >= 3 {
						if len(tokens) == 3 {
							value := convertValue(tokens[2])
							filterList = append(filterList, []any{tokens[0], tokens[1], value})
						} else {
							valueStr := strings.Join(tokens[2:], " ")
							value := convertValue(valueStr)
							filterList = append(filterList, []any{tokens[0], tokens[1], value})
						}
					}
				}

				if len(filterList) != tc.expectedCount {
					t.Errorf("Expected %d filters, got %d for '%s'", tc.expectedCount, len(filterList), searchText)
				}
			}
		})
	}
}

func TestMksmartCommand(t *testing.T) {
	client = &mockClient{}

	// Test mksmart command suggestions
	doc := createTestDocument("mks")
	suggests := completer(doc)
	if len(suggests) != 1 || suggests[0].Text != "mksmart" {
		t.Errorf("Expected 1 'mksmart' suggestion for 'mks', got %v", suggests)
	}

	// Test mksmart with simple filter
	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Valid simple filter",
			args:        []string{"Facturas_2025", "title", "match", "comprovativo"},
			expectError: false,
		},
		{
			name:        "Valid complex filter",
			args:        []string{"Documents", "title", "==", "Document,owner", "~=", "admin"},
			expectError: false,
		},
		{
			name:        "Missing arguments",
			args:        []string{"FolderName"},
			expectError: true,
		},
		{
			name:        "Empty arguments",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't easily test the full execution without output capture,
			// but we can at least verify the command doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("mksmart command panicked: %v", r)
				}
			}()

			// This will execute the command through our mock client
			mksmart(tc.args)
		})
	}
}

type enhancedMockClient struct {
	evaluateCalled bool
	listCalled     bool
}

func (c *enhancedMockClient) Login() error {
	return nil
}

func (c *enhancedMockClient) GetCurrentUser() (*antbox.User, error) {
	return &antbox.User{
		Email:  "enhanced@example.com",
		Name:   "Enhanced User",
		Groups: []string{"enhanced"},
	}, nil
}

func (c *enhancedMockClient) GetNode(uuid string) (*antbox.Node, error) {
	if uuid == "smartfolder-uuid" {
		return &antbox.Node{
			UUID:     "smartfolder-uuid",
			Title:    "Test Smart Folder",
			Mimetype: "application/vnd.antbox.smartfolder",
			Parent:   "--root--",
		}, nil
	} else if uuid == "regular-folder-uuid" {
		return &antbox.Node{
			UUID:     "regular-folder-uuid",
			Title:    "Test Regular Folder",
			Mimetype: "application/vnd.antbox.folder",
			Parent:   "--root--",
		}, nil
	} else if uuid == "--root--" {
		return &antbox.Node{UUID: "--root--", Title: "root", Parent: ""}, nil
	}
	return &antbox.Node{UUID: "test-uuid", Title: "test-title", Parent: "--root--", Mimetype: "application/pdf"}, nil
}

func (c *enhancedMockClient) ListNodes(parent string) ([]antbox.Node, error) {
	c.listCalled = true
	return []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}, nil
}

func (c *enhancedMockClient) CreateFolder(parent, name string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.folder"}, nil
}

func (c *enhancedMockClient) CreateSmartFolder(parent, name string, filters antbox.NodeFilters) (*antbox.Node, error) {
	return &antbox.Node{UUID: "smart-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.smartfolder"}, nil
}

func (c *enhancedMockClient) SetAuthHeader(req *http.Request) {}

func (c *enhancedMockClient) RemoveNode(uuid string) error {
	return nil
}

func (c *enhancedMockClient) MoveNode(uuid, newParent string) error {
	return nil
}

func (c *enhancedMockClient) ChangeNodeName(uuid, newName string) error {
	return nil
}

func (c *enhancedMockClient) CreateFile(filePath string, metadata antbox.NodeCreate) (*antbox.Node, error) {
	parentUuid := metadata.Parent
	return &antbox.Node{UUID: "uploaded-uuid", Title: "uploaded-file.txt", Parent: parentUuid}, nil
}

func (c *enhancedMockClient) CreateNode(node antbox.NodeCreate) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-node-uuid", Title: node.Title, Parent: node.Parent, Mimetype: node.Mimetype}, nil
}

func (c *enhancedMockClient) UpdateNode(uuid string, metadata antbox.NodeUpdate) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-title"}, nil
}

func (c *enhancedMockClient) UpdateFile(uuid, filePath string) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-file.txt", Parent: "--root--"}, nil
}

func (c *enhancedMockClient) FindNodes(filters antbox.NodeFilters, pageSize, pageToken int) (*antbox.NodeFilterResult, error) {
	return &antbox.NodeFilterResult{
		Nodes:     []antbox.Node{{UUID: "found-uuid", Title: "found-node", Mimetype: "text/plain"}},
		PageSize:  pageSize,
		PageToken: pageToken,
	}, nil
}

func (c *enhancedMockClient) EvaluateNode(uuid string) ([]antbox.Node, error) {
	c.evaluateCalled = true
	return []antbox.Node{
		{UUID: "eval-node-1", Title: "Evaluated Node 1", Mimetype: "text/plain"},
		{UUID: "eval-node-2", Title: "Evaluated Node 2", Mimetype: "application/pdf"},
	}, nil
}

func (c *enhancedMockClient) DownloadNode(uuid, downloadPath string) error {
	return nil
}

func (c *enhancedMockClient) GetBreadcrumbs(uuid string) ([]antbox.Node, error) {
	return []antbox.Node{
		{UUID: "--root--", Title: "root", Parent: ""},
		{UUID: "test-uuid", Title: "test-title", Parent: "--root--"},
	}, nil
}

func (c *enhancedMockClient) ChatWithAgent(agentUUID string, message string, conversationID string, temperature *float64, maxTokens *int, history []map[string]any) (antbox.ChatHistory, error) {
	text := "Mock chat response"
	return antbox.ChatHistory{
		{
			Role: antbox.ChatMessageRoleModel,
			Parts: []antbox.ChatMessagePart{
				{Text: &text},
			},
		},
	}, nil
}

func (c *enhancedMockClient) AnswerFromAgent(agentUUID string, query string, temperature *float64, maxTokens *int) (antbox.ChatHistory, error) {
	text := "Mock answer response"
	return antbox.ChatHistory{
		{
			Role: antbox.ChatMessageRoleModel,
			Parts: []antbox.ChatMessagePart{
				{Text: &text},
			},
		},
	}, nil
}

// New interface methods
func (c *enhancedMockClient) CopyNode(uuid, parent, title string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "copied-uuid", Title: title, Parent: parent}, nil
}

func (c *enhancedMockClient) CloneNode(uuid string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "cloned-uuid", Title: "Clone of test-title", Parent: "--root--"}, nil
}

func (c *enhancedMockClient) ExportNode(uuid string, format string) ([]byte, error) {
	return []byte("exported content"), nil
}

func (c *enhancedMockClient) ListAgents() ([]antbox.Agent, error) {
	return []antbox.Agent{{UUID: "agent-uuid", Name: "Test Agent"}}, nil
}

func (c *enhancedMockClient) UploadAgent(filePath string) (*antbox.Agent, error) {
	return &antbox.Agent{UUID: "agent-uuid", Name: "test-agent"}, nil
}

func (c *enhancedMockClient) GetAgent(uuid string) (*antbox.Agent, error) {
	return &antbox.Agent{UUID: uuid, Name: "Test Agent"}, nil
}

func (c *enhancedMockClient) DeleteAgent(uuid string) error {
	return nil
}

func (c *enhancedMockClient) ListAPIKeys() ([]antbox.APIKey, error) {
	return []antbox.APIKey{{UUID: "api-key-uuid", Description: "Test API Key"}}, nil
}

func (c *enhancedMockClient) CreateAPIKey(request antbox.APIKeyCreate) (*antbox.APIKey, error) {
	return &antbox.APIKey{UUID: "new-api-key-uuid", Description: request.Description}, nil
}

func (c *enhancedMockClient) GetAPIKey(uuid string) (*antbox.APIKey, error) {
	return &antbox.APIKey{UUID: uuid, Description: "Test API Key"}, nil
}

func (c *enhancedMockClient) DeleteAPIKey(uuid string) error {
	return nil
}

func (c *enhancedMockClient) ListUsers() ([]antbox.User, error) {
	return []antbox.User{{UUID: "user-uuid", Email: "test@example.com"}}, nil
}

func (c *enhancedMockClient) CreateUser(user antbox.UserCreate) (*antbox.User, error) {
	return &antbox.User{UUID: "new-user-uuid", Email: user.Email}, nil
}

func (c *enhancedMockClient) GetUser(email string) (*antbox.User, error) {
	return &antbox.User{UUID: "user-uuid", Email: email}, nil
}

func (c *enhancedMockClient) UpdateUser(email string, user antbox.UserUpdate) (*antbox.User, error) {
	return &antbox.User{UUID: "user-uuid", Email: email}, nil
}

func (c *enhancedMockClient) DeleteUser(uuid string) error {
	return nil
}

func (c *enhancedMockClient) ListGroups() ([]antbox.Group, error) {
	return []antbox.Group{{UUID: "group-uuid", Title: "Test Group"}}, nil
}

func (c *enhancedMockClient) CreateGroup(group antbox.GroupCreate) (*antbox.Group, error) {
	return &antbox.Group{UUID: "new-group-uuid", Title: group.Title}, nil
}

func (c *enhancedMockClient) GetGroup(uuid string) (*antbox.Group, error) {
	return &antbox.Group{UUID: uuid, Title: "Test Group"}, nil
}

func (c *enhancedMockClient) DeleteGroup(uuid string) error {
	return nil
}

func (c *enhancedMockClient) ListDocs() ([]antbox.DocInfo, error) {
	return []antbox.DocInfo{{UUID: "doc-uuid", Description: "Test Documentation"}}, nil
}

func (c *enhancedMockClient) GetDoc(uuid string) (string, error) {
	return "# Test Documentation\n\nThis is test documentation content.", nil
}

func (c *enhancedMockClient) GetAuditLog(uuid, mimetype string) ([]antbox.AuditEvent, error) {
	return (&mockClient{}).GetAuditLog(uuid, mimetype)
}

func TestSmartfolderCdAndLsBehavior(t *testing.T) {
	originalClient := client
	defer func() { client = originalClient }()

	// Test cd to smartfolder - should use evaluate endpoint
	t.Run("cd to smartfolder uses evaluate", func(t *testing.T) {
		mockClient := &enhancedMockClient{}
		client = mockClient

		cd([]string{"smartfolder-uuid"})

		if !mockClient.evaluateCalled {
			t.Errorf("Expected EvaluateNode to be called for smartfolder, but it wasn't")
		}
		if mockClient.listCalled {
			t.Errorf("Expected ListNodes NOT to be called for smartfolder, but it was")
		}
	})

	// Test cd to regular folder - should use list endpoint
	t.Run("cd to regular folder uses list", func(t *testing.T) {
		mockClient := &enhancedMockClient{}
		client = mockClient

		cd([]string{"regular-folder-uuid"})

		if mockClient.evaluateCalled {
			t.Errorf("Expected EvaluateNode NOT to be called for regular folder, but it was")
		}
		if !mockClient.listCalled {
			t.Errorf("Expected ListNodes to be called for regular folder, but it wasn't")
		}
	})

	// Test ls with smartfolder as current folder
	t.Run("ls in smartfolder uses evaluate", func(t *testing.T) {
		mockClient := &enhancedMockClient{}
		client = mockClient

		// Set current node to smartfolder
		originalCurrentNode := currentNode
		currentNode = antbox.Node{
			UUID:     "smartfolder-uuid",
			Title:    "Smart Folder",
			Mimetype: "application/vnd.antbox.smartfolder",
		}
		defer func() { currentNode = originalCurrentNode }()

		ls([]string{})

		if !mockClient.evaluateCalled {
			t.Errorf("Expected EvaluateNode to be called when ls in smartfolder, but it wasn't")
		}
		if mockClient.listCalled {
			t.Errorf("Expected ListNodes NOT to be called when ls in smartfolder, but it was")
		}
	})

	// Test ls with regular folder as current folder
	t.Run("ls in regular folder uses list", func(t *testing.T) {
		mockClient := &enhancedMockClient{}
		client = mockClient

		// Set current node to regular folder
		originalCurrentNode := currentNode
		currentNode = antbox.Node{
			UUID:     "regular-folder-uuid",
			Title:    "Regular Folder",
			Mimetype: "application/vnd.antbox.folder",
		}
		defer func() { currentNode = originalCurrentNode }()

		ls([]string{})

		if mockClient.evaluateCalled {
			t.Errorf("Expected EvaluateNode NOT to be called when ls in regular folder, but it was")
		}
		if !mockClient.listCalled {
			t.Errorf("Expected ListNodes to be called when ls in regular folder, but it wasn't")
		}
	})

	// Test ls with explicit smartfolder UUID as argument
	t.Run("ls with smartfolder UUID uses evaluate", func(t *testing.T) {
		mockClient := &enhancedMockClient{}
		client = mockClient

		ls([]string{"smartfolder-uuid"})

		if !mockClient.evaluateCalled {
			t.Errorf("Expected EvaluateNode to be called when ls with smartfolder UUID, but it wasn't")
		}
		if mockClient.listCalled {
			t.Errorf("Expected ListNodes NOT to be called when ls with smartfolder UUID, but it was")
		}
	})
}

func TestCommandExecuteCoverageForNonInteractiveCommands(t *testing.T) {
	client = &mockClient{}
	currentNode = antbox.Node{UUID: "--root--", Title: "root", Mimetype: "application/vnd.antbox.folder"}
	currentNodes = []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}
	cachedAgents = []antbox.Agent{{UUID: "agent-uuid", Name: "Test Agent"}}

	tests := []struct {
		name     string
		fn       func()
		contains string
	}{
		{"agents", func() { (&AgentsCommand{}).Execute(nil) }, "Available agents"},
		{"clone", func() { (&CloneCommand{}).Execute([]string{"test-uuid"}) }, "Node cloned successfully"},
		{"find", func() { (&FindCommand{}).Execute([]string{"test"}) }, "Found"},
		{"whoami", func() { (&WhoAmICommand{}).Execute(nil) }, "Current user"},
		{"help", func() { (&HelpCommand{}).Execute(nil) }, "Agent shortcuts"},
		{"status", func() { (&StatusCommand{}).Execute(nil) }, "Cached Resource Statistics"},
		{"reload", func() { (&ReloadCommand{}).Execute(nil) }, "Successfully reloaded"},
		{"docs", func() { (&DocsCommand{}).Execute(nil) }, "Available documents"},
		{"audit", func() { (&AuditCommand{}).Execute([]string{"test-uuid"}) }, "NodeUpdatedEvent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(t, tt.fn)
			if !strings.Contains(output, tt.contains) {
				t.Fatalf("expected output to contain %q, got %q", tt.contains, output)
			}
		})
	}
}

func TestAuditCommandHidesPayloadUnlessVerbose(t *testing.T) {
	client = &mockClient{}

	output := captureOutput(t, func() { (&AuditCommand{}).Execute([]string{"test-uuid"}) })
	if strings.Contains(output, "Payload") || strings.Contains(output, "Old") || strings.Contains(output, "deleted") {
		t.Fatalf("expected default audit output to hide payload, got %q", output)
	}
	if !strings.Contains(output, "#1     NodeUpdatedEvent") || !strings.Contains(output, "#2     NodeDeletedEvent") {
		t.Fatalf("expected audit event summary lines, got %q", output)
	}
	if strings.Index(output, "#2     NodeDeletedEvent") > strings.Index(output, "#1     NodeUpdatedEvent") {
		t.Fatalf("expected audit events sorted by descending sequence, got %q", output)
	}

	verboseOutput := captureOutput(t, func() { (&AuditCommand{}).Execute([]string{"-v", "test-uuid"}) })
	if !strings.Contains(verboseOutput, "Payload:") || !strings.Contains(verboseOutput, "Old") || !strings.Contains(verboseOutput, "deleted") {
		t.Fatalf("expected verbose audit output to include payload, got %q", verboseOutput)
	}
}

func TestGetCommandsReturnsRegistry(t *testing.T) {
	registered := GetCommands()
	if registered["ls"] == nil || registered["agents"] == nil {
		t.Fatalf("expected command registry to include ls and agents")
	}
}

func TestStatJoinPermissions(t *testing.T) {
	joined := joinPermissions([]antbox.Permission{antbox.PermissionRead, antbox.PermissionWrite})
	if joined != "Read, Write" {
		t.Fatalf("unexpected joined permissions: %q", joined)
	}
}

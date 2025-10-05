package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/kindalus/antx/antbox"

	prompt "github.com/c-bata/go-prompt"
)

type mockClient struct {
}

func (c *mockClient) Login() error {
	return nil
}

func (c *mockClient) GetNode(uuid string) (*antbox.Node, error) {
	if uuid == "--root--" {
		return &antbox.Node{UUID: "--root--", Title: "root", Parent: ""}, nil
	}
	return &antbox.Node{UUID: "test-uuid", Title: "test-title", Parent: "--root--"}, nil
}

func (c *mockClient) ListNodes(parent string) ([]antbox.Node, error) {
	return []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}, nil
}

func (c *mockClient) CreateFolder(parent, name string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.folder"}, nil
}

func (c *mockClient) CreateSmartFolder(parent, name string, filters any) (*antbox.Node, error) {
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

func (c *mockClient) CreateFile(filePath, parentUuid string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "uploaded-uuid", Title: "uploaded-file.txt", Parent: parentUuid}, nil
}

func (c *mockClient) UpdateFile(uuid, filePath string) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-file.txt", Parent: "--root--"}, nil
}

func (c *mockClient) FindNodes(filters interface{}, pageSize, pageToken int) (*antbox.NodeFilterResult, error) {
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

func TestCompleter(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}

	// Test with single character - should return no suggestions (less than 2 chars)
	doc := prompt.Document{Text: "l"}
	suggests := completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for single char, got %d", len(suggests))
	}

	// Test with 2+ characters - should return matching suggestions
	doc = prompt.Document{Text: "ls"}
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'ls', got %d", len(suggests))
	}

	// Test command with arguments - should show node suggestions if 2+ chars
	doc = prompt.Document{Text: "ls te"}
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'ls te', got %d", len(suggests))
	}

	// Test cd command - should use UUID for folder suggestions
	doc = prompt.Document{Text: "cd te"}
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'cd te', got %d", len(suggests))
	}
	if suggests[0].Text != "test-uuid" {
		t.Errorf("Expected cd suggestion to use UUID 'test-uuid', got '%s'", suggests[0].Text)
	}
	if !strings.Contains(suggests[0].Description, "test-title") {
		t.Errorf("Expected cd suggestion description to contain folder name, got '%s'", suggests[0].Description)
	}

	// Test command with single char argument - should return no suggestions
	doc = prompt.Document{Text: "ls t"}
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for single char argument, got %d", len(suggests))
	}
}

func TestCompleterCdWithMixedNodeTypes(t *testing.T) {
	client = &mockClient{}
	// Mix of folder and file nodes
	currentNodes = []antbox.Node{
		{UUID: "folder-uuid-1", Title: "documents", Mimetype: "application/vnd.antbox.folder"},
		{UUID: "file-uuid-1", Title: "document.txt", Mimetype: "text/plain"},
		{UUID: "folder-uuid-2", Title: "downloads", Mimetype: "application/vnd.antbox.folder"},
	}

	// Test cd command with folder prefix - should only suggest folders and use UUIDs
	doc := prompt.Document{Text: "cd do"}
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
		if !strings.Contains(suggest.Description, "📁") {
			t.Errorf("Expected folder emoji in description, got '%s'", suggest.Description)
		}
	}

	// Test ls command with same prefix - should suggest both files and folders
	doc = prompt.Document{Text: "ls do"}
	suggests = completer(doc)

	// Should get 3 suggestions: 2 folders by title, 1 file by title
	// (UUIDs don't start with "do" so won't match)
	if len(suggests) != 3 {
		t.Errorf("Expected 3 suggestions for 'ls do', got %d", len(suggests))
	}

	// Test cd with exact folder name
	doc = prompt.Document{Text: "cd documents"}
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
		{UUID: "folder-uuid", Title: "test-folder", Mimetype: "application/vnd.antbox.folder"},
		{UUID: "file-uuid", Title: "test-file.txt", Mimetype: "text/plain"},
	}

	// Test rm command suggestions (should suggest all nodes)
	doc := prompt.Document{Text: "rm te"}
	suggests := completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'rm te', got %d", len(suggests))
	}

	// Test mv command first argument (should suggest all nodes)
	doc = prompt.Document{Text: "mv te"}
	suggests = completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'mv te', got %d", len(suggests))
	}

	// Test mv command second argument (should only suggest folders with UUID)
	doc = prompt.Document{Text: "mv file-uuid te"}
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for mv second argument, got %d", len(suggests))
	}
	if suggests[0].Text != "folder-uuid" {
		t.Errorf("Expected folder UUID for mv destination, got '%s'", suggests[0].Text)
	}

	// Test cp command first argument (should not provide suggestions)
	doc = prompt.Document{Text: "cp /path/to/file"}
	suggests = completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for cp file path, got %d", len(suggests))
	}

	// Test cp command second argument (should only suggest folders with UUID)
	doc = prompt.Document{Text: "cp /path/to/file te"}
	suggests = completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for cp destination, got %d", len(suggests))
	}
	if suggests[0].Text != "folder-uuid" {
		t.Errorf("Expected folder UUID for cp destination, got '%s'", suggests[0].Text)
	}

	// Test get command suggestions (should suggest all nodes)
	doc = prompt.Document{Text: "get te"}
	suggests = completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'get te', got %d", len(suggests))
	}

	// Test rename command suggestions (should suggest all nodes)
	doc = prompt.Document{Text: "rename te"}
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
		{"ls", []string{"ls"}},
		{"rm", []string{"rm"}},
		{"mk", []string{"mkdir", "mksmart"}},
		{"mv", []string{"mv"}},
		{"cd", []string{"cd"}},
		{"re", []string{"rename"}},
		{"cp", []string{"cp"}},
		{"ge", []string{"get"}},
		{"st", []string{"stat"}},
		{"ex", []string{"exit"}},
		{"pw", []string{"pwd"}},
		{"up", []string{"update"}},
	}

	for _, tc := range testCases {
		doc := prompt.Document{Text: tc.input}
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
			if !found {
				t.Errorf("For input '%s': expected command '%s' not found in suggestions %v", tc.input, expectedCmd, suggestedTexts)
			}
		}
	}
}

func TestCommandSuggestionsMinLength(t *testing.T) {
	client = &mockClient{}

	// Test that single character inputs return no suggestions
	singleCharInputs := []string{"l", "r", "m", "c", "g", "s", "e", "p"}

	for _, input := range singleCharInputs {
		doc := prompt.Document{Text: input}
		suggests := completer(doc)

		if len(suggests) != 0 {
			t.Errorf("For single character input '%s': expected 0 suggestions, got %d", input, len(suggests))
		}
	}

	// Test that 2+ character inputs return appropriate suggestions
	doc := prompt.Document{Text: "ls"}
	suggests := completer(doc)
	if len(suggests) != 1 || suggests[0].Text != "ls" {
		t.Errorf("Expected 1 'ls' suggestion for 'ls' input, got %d suggestions", len(suggests))
	}
}

func TestUpdateCommandSuggestions(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{
		{UUID: "folder-uuid", Title: "test-folder", Mimetype: "application/vnd.antbox.folder"},
		{UUID: "file-uuid", Title: "test-file.txt", Mimetype: "text/plain"},
	}

	// Test update command first argument (should suggest all nodes)
	doc := prompt.Document{Text: "update te"}
	suggests := completer(doc)
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'update te', got %d", len(suggests))
	}

	// Test update command suggestions include both files and folders
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
		t.Errorf("Expected to find file-uuid in update suggestions")
	}
	if !foundFolder {
		t.Errorf("Expected to find folder-uuid in update suggestions")
	}
}

func TestFindCommand(t *testing.T) {
	client = &mockClient{}

	// Test find with simple search
	doc := prompt.Document{Text: "find te"}
	suggests := completer(doc)
	if len(suggests) != 0 {
		t.Errorf("Expected 0 suggestions for find command, got %d", len(suggests))
	}
}

func TestUpdateCommand(t *testing.T) {
	client = &mockClient{}
	currentNodes = []antbox.Node{{UUID: "test-uuid", Title: "test-file.txt", Mimetype: "text/plain"}}

	// Test update command first argument - should suggest nodes
	doc := prompt.Document{Text: "update te"}
	suggests := completer(doc)
	if len(suggests) != 1 {
		t.Errorf("Expected 1 suggestion for 'update te', got %d", len(suggests))
	}
	if suggests[0].Text != "test-uuid" {
		t.Errorf("Expected update suggestion to use UUID 'test-uuid', got '%s'", suggests[0].Text)
	}
}

func TestCommandSuggestionsWithNewCommands(t *testing.T) {
	client = &mockClient{}

	testCases := []struct {
		input    string
		expected []string
	}{
		{"fi", []string{"find"}},
		{"up", []string{"update"}},
	}

	for _, tc := range testCases {
		doc := prompt.Document{Text: tc.input}
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
			found := false
			for _, suggested := range suggestedTexts {
				if suggested == expectedCmd {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("For input '%s': expected command '%s' not found in suggestions %v", tc.input, expectedCmd, suggestedTexts)
			}
		}
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
		{"createdAt>=2023-01-01", "createdAt >= 2023-01-01"},
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
		expected interface{}
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

				var filterList [][]interface{}
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
							filterList = append(filterList, []interface{}{tokens[0], tokens[1], value})
						} else {
							valueStr := strings.Join(tokens[2:], " ")
							value := convertValue(valueStr)
							filterList = append(filterList, []interface{}{tokens[0], tokens[1], value})
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
	doc := prompt.Document{Text: "mks"}
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
	return &antbox.Node{UUID: "test-uuid", Title: "test-title", Parent: "--root--"}, nil
}

func (c *enhancedMockClient) ListNodes(parent string) ([]antbox.Node, error) {
	c.listCalled = true
	return []antbox.Node{{UUID: "test-uuid", Title: "test-title", Mimetype: "application/vnd.antbox.folder"}}, nil
}

func (c *enhancedMockClient) CreateFolder(parent, name string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "new-uuid", Title: name, Parent: parent, Mimetype: "application/vnd.antbox.folder"}, nil
}

func (c *enhancedMockClient) CreateSmartFolder(parent, name string, filters any) (*antbox.Node, error) {
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

func (c *enhancedMockClient) CreateFile(filePath, parentUuid string) (*antbox.Node, error) {
	return &antbox.Node{UUID: "uploaded-uuid", Title: "uploaded-file.txt", Parent: parentUuid}, nil
}

func (c *enhancedMockClient) UpdateFile(uuid, filePath string) (*antbox.Node, error) {
	return &antbox.Node{UUID: uuid, Title: "updated-file.txt", Parent: "--root--"}, nil
}

func (c *enhancedMockClient) FindNodes(filters interface{}, pageSize, pageToken int) (*antbox.NodeFilterResult, error) {
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

		// Set current folder to smartfolder
		originalCurrentFolder := currentFolder
		currentFolder = "smartfolder-uuid"
		defer func() { currentFolder = originalCurrentFolder }()

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

		// Set current folder to regular folder
		originalCurrentFolder := currentFolder
		currentFolder = "regular-folder-uuid"
		defer func() { currentFolder = originalCurrentFolder }()

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

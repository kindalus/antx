package cli

import (
	"net/http"
	"strings"
	"testing"

	"antbox-cli/antbox"

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

func (c *mockClient) SetAuthHeader(req *http.Request) {}

func TestExecutor(t *testing.T) {
	client = &mockClient{}

	executor("ls")
	executor("pwd")
	executor("stat test-uuid")
	executor("cd test-uuid")
	executor("cd ..")
	executor("mkdir test-folder")
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
	if len(suggests) != 2 {
		t.Errorf("Expected 2 suggestions for 'ls te', got %d", len(suggests))
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

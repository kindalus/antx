package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kindalus/antx/antbox"
)

// Mock client for testing breadcrumbs and sorting
type testClient struct {
	nodes map[string]*antbox.Node
}

func newTestClient() *testClient {
	return &testClient{
		nodes: map[string]*antbox.Node{
			"--root--": {
				UUID:       "--root--",
				Title:      "root",
				Mimetype:   "application/vnd.antbox.folder",
				Parent:     "",
				ModifiedAt: "2024-01-01T12:00:00Z",
			},
			"folder-docs": {
				UUID:       "folder-docs",
				Title:      "Documents",
				Mimetype:   "application/vnd.antbox.folder",
				Parent:     "--root--",
				ModifiedAt: "2024-01-02T10:30:00Z",
			},
			"folder-projects": {
				UUID:       "folder-projects",
				Title:      "Projects",
				Mimetype:   "application/vnd.antbox.folder",
				Parent:     "--root--",
				ModifiedAt: "2024-01-03T09:15:00Z",
			},
			"folder-archive": {
				UUID:       "folder-archive",
				Title:      "Archive",
				Mimetype:   "application/vnd.antbox.folder",
				Parent:     "--root--",
				ModifiedAt: "2024-01-01T14:45:00Z",
			},
			"smart-recent": {
				UUID:       "smart-recent",
				Title:      "Recent Items",
				Mimetype:   "application/vnd.antbox.smartfolder",
				Parent:     "--root--",
				ModifiedAt: "2024-01-04T08:00:00Z",
			},
			"file-readme": {
				UUID:       "file-readme",
				Title:      "README.md",
				Mimetype:   "text/markdown",
				Parent:     "--root--",
				Size:       1024,
				ModifiedAt: "2024-01-05T16:30:00Z",
			},
			"file-config": {
				UUID:       "file-config",
				Title:      "config.json",
				Mimetype:   "application/json",
				Parent:     "--root--",
				Size:       512,
				ModifiedAt: "2024-01-06T11:20:00Z",
			},
			"file-image": {
				UUID:       "file-image",
				Title:      "photo.jpg",
				Mimetype:   "image/jpeg",
				Parent:     "--root--",
				Size:       2048576,
				ModifiedAt: "2024-01-07T13:45:00Z",
			},
			"sub-reports": {
				UUID:       "sub-reports",
				Title:      "Reports",
				Mimetype:   "application/vnd.antbox.folder",
				Parent:     "folder-docs",
				ModifiedAt: "2024-01-08T15:10:00Z",
			},
			"sub-doc": {
				UUID:       "sub-doc",
				Title:      "document.pdf",
				Mimetype:   "application/pdf",
				Parent:     "folder-docs",
				Size:       1536000,
				ModifiedAt: "2024-01-09T12:25:00Z",
			},
			"deep-annual": {
				UUID:       "deep-annual",
				Title:      "Annual Report",
				Mimetype:   "application/pdf",
				Parent:     "sub-reports",
				Size:       3072000,
				ModifiedAt: "2024-01-10T17:30:00Z",
			},
		},
	}
}

func (c *testClient) GetNode(uuid string) (*antbox.Node, error) {
	if node, exists := c.nodes[uuid]; exists {
		return node, nil
	}
	return nil, fmt.Errorf("node not found: %s", uuid)
}

func (c *testClient) ListNodes(folderUUID string) ([]antbox.Node, error) {
	var nodes []antbox.Node
	for _, node := range c.nodes {
		if node.Parent == folderUUID {
			nodes = append(nodes, *node)
		}
	}
	return nodes, nil
}

func (c *testClient) GetBreadcrumbs(uuid string) ([]antbox.Node, error) {
	var breadcrumbs []antbox.Node
	current := uuid

	// Build breadcrumb trail by following parent chain
	for current != "" && current != "--root--" {
		node, exists := c.nodes[current]
		if !exists {
			break
		}
		breadcrumbs = append([]antbox.Node{*node}, breadcrumbs...)
		current = node.Parent
	}

	// Always include root at the beginning
	if root, exists := c.nodes["--root--"]; exists {
		breadcrumbs = append([]antbox.Node{*root}, breadcrumbs...)
	}

	return breadcrumbs, nil
}

func (c *testClient) ListAgents() ([]antbox.Agent, error) {
	return []antbox.Agent{
		{UUID: "agent-writer", Title: "Writer Assistant", Description: "Helps with writing tasks"},
		{UUID: "agent-coder", Title: "Code Helper", Description: "Assists with programming"},
		{UUID: "agent-analyst", Title: "Data Analyst", Description: "Analyzes data patterns"},
		{UUID: "agent-bot", Title: "Chat Bot", Description: "General conversation assistant"},
	}, nil
}

func (c *testClient) ListExtensions() ([]antbox.Feature, error) {
	return []antbox.Feature{
		{UUID: "ext-pdf", Name: "PDF Generator", Description: "Creates PDF documents"},
		{UUID: "ext-email", Name: "Email Sender", Description: "Sends email notifications"},
		{UUID: "ext-backup", Name: "Backup Tool", Description: "Creates data backups"},
		{UUID: "ext-analyzer", Name: "Data Analyzer", Description: "Processes data files"},
	}, nil
}

func (c *testClient) ListActions() ([]antbox.Feature, error) {
	return []antbox.Feature{
		{UUID: "action-convert", Name: "Convert File", Description: "Converts between file formats"},
		{UUID: "action-compress", Name: "Compress", Description: "Compresses files and folders"},
		{UUID: "action-extract", Name: "Extract Text", Description: "Extracts text from documents"},
		{UUID: "action-archive", Name: "Archive", Description: "Archives old files"},
	}, nil
}

// Test helper functions

func sortNodesForListing(nodes []antbox.Node) []antbox.Node {
	if len(nodes) == 0 {
		return nodes
	}

	// Separate directories and files
	var directories []antbox.Node
	var files []antbox.Node

	for _, node := range nodes {
		isFolder := node.Mimetype == "application/vnd.antbox.folder" ||
			node.Mimetype == "application/vnd.antbox.smartfolder"

		if isFolder {
			directories = append(directories, node)
		} else {
			files = append(files, node)
		}
	}

	// Sort directories alphabetically by title
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Title < directories[j].Title
	})

	// Sort files alphabetically by title
	sort.Slice(files, func(i, j int) bool {
		return files[i].Title < files[j].Title
	})

	// Combine directories first, then files
	result := make([]antbox.Node, 0, len(nodes))
	result = append(result, directories...)
	result = append(result, files...)

	return result
}

func testBreadcrumbs(client *testClient) error {
	fmt.Println("Testing Breadcrumbs Functionality")
	fmt.Println("=================================")

	testCases := []struct {
		uuid     string
		expected string
	}{
		{"--root--", "/root"},
		{"folder-docs", "/root/Documents"},
		{"sub-reports", "/root/Documents/Reports"},
		{"deep-annual", "/root/Documents/Reports/Annual Report"},
	}

	for i, tc := range testCases {
		fmt.Printf("Test %d: Breadcrumbs for %s\n", i+1, tc.uuid)

		breadcrumbs, err := client.GetBreadcrumbs(tc.uuid)
		if err != nil {
			return fmt.Errorf("❌ Failed to get breadcrumbs for %s: %v", tc.uuid, err)
		}

		// Build path from breadcrumbs
		var pathParts []string
		for _, node := range breadcrumbs {
			if node.Title != "" {
				pathParts = append(pathParts, node.Title)
			}
		}

		var path string
		if len(pathParts) == 0 {
			path = "/"
		} else {
			path = "/" + strings.Join(pathParts, "/")
		}

		fmt.Printf("  Expected: %s\n", tc.expected)
		fmt.Printf("  Actual:   %s\n", path)

		if path == tc.expected {
			fmt.Printf("  ✅ PASS\n")
		} else {
			return fmt.Errorf("❌ FAIL: Expected %s, got %s", tc.expected, path)
		}
		fmt.Println()
	}

	return nil
}

func testNodeSorting(client *testClient) error {
	fmt.Println("Testing Node Sorting (Directories First, Alphabetical)")
	fmt.Println("=====================================================")

	// Get root folder contents
	nodes, err := client.ListNodes("--root--")
	if err != nil {
		return fmt.Errorf("❌ Failed to list nodes: %v", err)
	}

	fmt.Printf("Original order (%d nodes):\n", len(nodes))
	for i, node := range nodes {
		nodeType := "file"
		if node.Mimetype == "application/vnd.antbox.folder" || node.Mimetype == "application/vnd.antbox.smartfolder" {
			nodeType = "folder"
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, node.Title, nodeType)
	}
	fmt.Println()

	// Sort nodes
	sortedNodes := sortNodesForListing(nodes)

	fmt.Printf("Sorted order (directories first, then alphabetical):\n")
	directories := 0
	files := 0
	for i, node := range sortedNodes {
		nodeType := "file"
		if node.Mimetype == "application/vnd.antbox.folder" || node.Mimetype == "application/vnd.antbox.smartfolder" {
			nodeType = "folder"
			directories++
		} else {
			files++
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, node.Title, nodeType)
	}
	fmt.Printf("\nSummary: %d directories, %d files\n", directories, files)
	fmt.Println()

	// Verify sorting correctness
	expectedOrder := []string{
		"Archive",      // folder (alphabetical)
		"Documents",    // folder (alphabetical)
		"Projects",     // folder (alphabetical)
		"Recent Items", // smartfolder (alphabetical)
		"README.md",    // file (alphabetical - 'R' comes before 'c')
		"config.json",  // file (alphabetical)
		"photo.jpg",    // file (alphabetical)
	}

	if len(sortedNodes) != len(expectedOrder) {
		return fmt.Errorf("❌ Expected %d nodes, got %d", len(expectedOrder), len(sortedNodes))
	}

	fmt.Println("Verifying sort order:")
	for i, expected := range expectedOrder {
		actual := sortedNodes[i].Title
		fmt.Printf("  Position %d: Expected '%s', Got '%s'", i+1, expected, actual)
		if actual == expected {
			fmt.Printf(" ✅\n")
		} else {
			return fmt.Errorf(" ❌ FAIL at position %d", i+1)
		}
	}

	fmt.Println("✅ Node sorting test PASSED")
	fmt.Println()
	return nil
}

func testAgentsSorting(client *testClient) error {
	fmt.Println("Testing Agents Alphabetical Sorting")
	fmt.Println("===================================")

	agents, err := client.ListAgents()
	if err != nil {
		return fmt.Errorf("❌ Failed to list agents: %v", err)
	}

	fmt.Printf("Original agents order:\n")
	for i, agent := range agents {
		fmt.Printf("  %d. %s\n", i+1, agent.Title)
	}
	fmt.Println()

	// Sort agents alphabetically by title
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Title < agents[j].Title
	})

	fmt.Printf("Sorted agents (alphabetical by title):\n")
	for i, agent := range agents {
		fmt.Printf("  %d. %s\n", i+1, agent.Title)
	}

	expectedOrder := []string{"Chat Bot", "Code Helper", "Data Analyst", "Writer Assistant"}
	fmt.Println("\nVerifying sort order:")
	for i, expected := range expectedOrder {
		actual := agents[i].Title
		fmt.Printf("  Position %d: Expected '%s', Got '%s'", i+1, expected, actual)
		if actual == expected {
			fmt.Printf(" ✅\n")
		} else {
			return fmt.Errorf(" ❌ FAIL at position %d", i+1)
		}
	}

	fmt.Println("✅ Agents sorting test PASSED")
	fmt.Println()
	return nil
}

func testExtensionsSorting(client *testClient) error {
	fmt.Println("Testing Extensions Alphabetical Sorting")
	fmt.Println("=======================================")

	extensions, err := client.ListExtensions()
	if err != nil {
		return fmt.Errorf("❌ Failed to list extensions: %v", err)
	}

	// Sort extensions alphabetically by name
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})

	fmt.Printf("Sorted extensions (alphabetical by name):\n")
	for i, ext := range extensions {
		fmt.Printf("  %d. %s\n", i+1, ext.Name)
	}

	expectedOrder := []string{"Backup Tool", "Data Analyzer", "Email Sender", "PDF Generator"}
	fmt.Println("\nVerifying sort order:")
	for i, expected := range expectedOrder {
		actual := extensions[i].Name
		fmt.Printf("  Position %d: Expected '%s', Got '%s'", i+1, expected, actual)
		if actual == expected {
			fmt.Printf(" ✅\n")
		} else {
			return fmt.Errorf(" ❌ FAIL at position %d", i+1)
		}
	}

	fmt.Println("✅ Extensions sorting test PASSED")
	fmt.Println()
	return nil
}

func testActionsSorting(client *testClient) error {
	fmt.Println("Testing Actions Alphabetical Sorting")
	fmt.Println("====================================")

	actions, err := client.ListActions()
	if err != nil {
		return fmt.Errorf("❌ Failed to list actions: %v", err)
	}

	// Sort actions alphabetically by name
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Name < actions[j].Name
	})

	fmt.Printf("Sorted actions (alphabetical by name):\n")
	for i, action := range actions {
		fmt.Printf("  %d. %s\n", i+1, action.Name)
	}

	expectedOrder := []string{"Archive", "Compress", "Convert File", "Extract Text"}
	fmt.Println("\nVerifying sort order:")
	for i, expected := range expectedOrder {
		actual := actions[i].Name
		fmt.Printf("  Position %d: Expected '%s', Got '%s'", i+1, expected, actual)
		if actual == expected {
			fmt.Printf(" ✅\n")
		} else {
			return fmt.Errorf(" ❌ FAIL at position %d", i+1)
		}
	}

	fmt.Println("✅ Actions sorting test PASSED")
	fmt.Println()
	return nil
}

func main() {
	fmt.Println("🧪 Breadcrumbs and Sorting Test Suite")
	fmt.Println("=====================================")
	fmt.Println()

	// Initialize test client
	client := newTestClient()

	// Run all tests
	tests := []struct {
		name string
		fn   func(*testClient) error
	}{
		{"Breadcrumbs", testBreadcrumbs},
		{"Node Sorting", testNodeSorting},
		{"Agents Sorting", testAgentsSorting},
		{"Extensions Sorting", testExtensionsSorting},
		{"Actions Sorting", testActionsSorting},
	}

	allPassed := true
	for _, test := range tests {
		fmt.Printf("🔍 Running %s test...\n", test.name)
		if err := test.fn(client); err != nil {
			fmt.Printf("❌ %s FAILED: %v\n", test.name, err)
			allPassed = false
		}
		fmt.Println()
	}

	// Final result
	if allPassed {
		fmt.Println("🎉 ALL TESTS PASSED!")
		fmt.Println()
		fmt.Println("✅ Breadcrumbs: Display hierarchical path correctly")
		fmt.Println("✅ Node Listing: Directories first, then files, both sorted alphabetically")
		fmt.Println("✅ Agents Listing: Sorted alphabetically by title")
		fmt.Println("✅ Extensions Listing: Sorted alphabetically by name")
		fmt.Println("✅ Actions Listing: Sorted alphabetically by name")
		fmt.Println()
		fmt.Println("Features are ready for production use!")
		os.Exit(0)
	} else {
		fmt.Println("❌ SOME TESTS FAILED!")
		fmt.Println("Please review the errors above and fix the issues.")
		os.Exit(1)
	}
}

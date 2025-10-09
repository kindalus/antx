package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

// Simulate CLI state for testing
var (
	currentNode  antbox.Node
	cliHistory   []string
	testCommands []string
	commandIndex int
)

// Mock client for testing
type testClient struct{}

func (c *testClient) GetNode(uuid string) (*antbox.Node, error) {
	return &antbox.Node{
		UUID:     uuid,
		Title:    "Test Node",
		Mimetype: "application/vnd.antbox.folder",
		Parent:   "--root--",
	}, nil
}

func (c *testClient) ListNodes(parent string) ([]antbox.Node, error) {
	return []antbox.Node{}, nil
}

// Test history functions
func createTestHistory() []string {
	return []string{
		"ls",
		"cd documents-uuid",
		"stat .",
		"ls",
		"cd projects-uuid",
		"run action-123 .",
		"stat ..",
		"cd ..",
		"find title match document",
		"upload file.txt",
	}
}

func saveTestConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".antx_test")

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write current node UUID
	fmt.Fprintf(file, "projects-uuid\n")

	// Write blank line
	fmt.Fprintf(file, "\n")

	// Write test history
	for _, cmd := range createTestHistory() {
		fmt.Fprintf(file, "%s\n", cmd)
	}

	return nil
}

func loadTestHistory() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".antx_test")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var history []string

	// Skip first line (UUID) and second line (blank)
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			history = append(history, line)
		}
	}

	return history, nil
}

func cleanupTestConfig() {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".antx_test")
	os.Remove(configPath)
}

// Test executor that simulates command execution
func testExecutor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	fmt.Printf("Executed: %s\n", input)

	// Add to history (simulate real CLI behavior)
	if len(cliHistory) == 0 || cliHistory[len(cliHistory)-1] != input {
		cliHistory = append(cliHistory, input)
	}

	// Handle special commands
	switch input {
	case "exit", "quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "history":
		fmt.Println("\nCurrent History:")
		for i, cmd := range cliHistory {
			fmt.Printf("  %d  %s\n", i+1, cmd)
		}
		fmt.Printf("Total: %d commands\n", len(cliHistory))
	case "save":
		if err := saveTestConfig(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
		} else {
			fmt.Println("Configuration saved!")
		}
	case "demo":
		runDemo()
	default:
		// Simulate command execution
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println()
}

// Test completer (minimal for testing)
func testCompleter(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{
		{Text: "ls", Description: "List contents"},
		{Text: "cd", Description: "Change directory"},
		{Text: "stat", Description: "Show node info"},
		{Text: "history", Description: "Show command history"},
		{Text: "demo", Description: "Run demo commands"},
		{Text: "save", Description: "Save current state"},
		{Text: "exit", Description: "Exit the test"},
	}

	text := d.GetWordBeforeCursor()
	if text == "" {
		return suggestions
	}

	var filtered []prompt.Suggest
	for _, s := range suggestions {
		if strings.HasPrefix(s.Text, text) {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

func runDemo() {
	fmt.Println("Running demo commands...")

	demoCommands := []string{
		"ls",
		"cd documents-uuid",
		"stat .",
		"cd projects-uuid",
		"ls",
	}

	for _, cmd := range demoCommands {
		fmt.Printf("Auto-executing: %s\n", cmd)
		testExecutor(cmd)
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Demo complete! Try using up/down arrows to navigate history.")
}

func main() {
	fmt.Println("ANTX CLI History Integration Test")
	fmt.Println("=================================")
	fmt.Println()

	// Initialize current node
	currentNode = antbox.Node{
		UUID:     "--root--",
		Title:    "root",
		Mimetype: "application/vnd.antbox.folder",
	}

	// Load test history
	var err error
	cliHistory, err = loadTestHistory()
	if err != nil {
		fmt.Printf("Warning: Could not load test history: %v\n", err)
		cliHistory = []string{}
	}

	if len(cliHistory) == 0 {
		fmt.Println("No previous history found. Creating test history...")
		cliHistory = createTestHistory()
		if err := saveTestConfig(); err != nil {
			fmt.Printf("Error saving initial config: %v\n", err)
		}
	}

	fmt.Printf("Loaded %d commands from history\n", len(cliHistory))
	fmt.Println()

	fmt.Println("History Integration Features:")
	fmt.Println("-----------------------------")
	fmt.Println("✓ Up/Down arrows navigate through saved history")
	fmt.Println("✓ History persists across sessions")
	fmt.Println("✓ Commands are automatically added to history")
	fmt.Println("✓ History is integrated with go-prompt")
	fmt.Println()

	fmt.Println("Test Commands:")
	fmt.Println("--------------")
	fmt.Println("  history  - Show current command history")
	fmt.Println("  demo     - Run demo commands automatically")
	fmt.Println("  save     - Save current state to config file")
	fmt.Println("  exit     - Exit the test")
	fmt.Println()

	fmt.Println("Instructions:")
	fmt.Println("-------------")
	fmt.Println("1. Use UP/DOWN arrows to navigate through history")
	fmt.Println("2. Type new commands to add them to history")
	fmt.Println("3. Type 'history' to see current command list")
	fmt.Println("4. Type 'demo' to auto-execute some test commands")
	fmt.Println("5. Type 'exit' to quit")
	fmt.Println()

	// Create prompt with history integration
	p := prompt.New(
		testExecutor,
		testCompleter,
		prompt.OptionTitle("ANTX CLI History Test"),
		prompt.OptionPrefix("test # "),
		prompt.OptionHistory(cliHistory), // This integrates our saved history!
		prompt.OptionMaxSuggestion(8),
		prompt.OptionCompletionWordSeparator(" "),
		// Add some helpful key bindings
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Println("\nUse 'exit' to quit properly")
			},
		}),
	)

	// Cleanup function
	defer func() {
		fmt.Println("\nCleaning up test files...")
		cleanupTestConfig()
		fmt.Println("Test completed!")
	}()

	// Run the prompt
	fmt.Println("Starting interactive test (try UP/DOWN arrows!):")
	fmt.Println(strings.Repeat("=", 50))
	p.Run()
}

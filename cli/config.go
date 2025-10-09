package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kindalus/antx/antbox"
)

const (
	configFileName = ".antx"
	maxHistorySize = 20
)

// CLIConfig holds the persistent CLI state
type CLIConfig struct {
	CurrentNodeUUID string
	History         []string
}

// getConfigDir returns the user's configuration directory
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return homeDir, nil
}

// getConfigFilePath returns the full path to the .antx config file
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// saveConfig saves the current CLI state to disk
func saveConfig(config *CLIConfig) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %v", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	// Write current node UUID on first line
	if _, err := fmt.Fprintf(file, "%s\n", config.CurrentNodeUUID); err != nil {
		return fmt.Errorf("failed to write current node UUID: %v", err)
	}

	// Write blank line
	if _, err := fmt.Fprintf(file, "\n"); err != nil {
		return fmt.Errorf("failed to write blank line: %v", err)
	}

	// Write history (last 20 commands)
	historyToSave := config.History
	if len(historyToSave) > maxHistorySize {
		historyToSave = historyToSave[len(historyToSave)-maxHistorySize:]
	}

	for _, cmd := range historyToSave {
		if _, err := fmt.Fprintf(file, "%s\n", cmd); err != nil {
			return fmt.Errorf("failed to write history entry: %v", err)
		}
	}

	return nil
}

// loadConfig loads the CLI state from disk
func loadConfig() (*CLIConfig, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %v", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &CLIConfig{
			CurrentNodeUUID: "--root--",
			History:         []string{},
		}, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	config := &CLIConfig{
		CurrentNodeUUID: "--root--",
		History:         []string{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if lineNum == 1 {
			// First line: current node UUID
			if line != "" {
				config.CurrentNodeUUID = line
			}
		} else if lineNum == 2 {
			// Second line: blank line (skip)
			continue
		} else {
			// Remaining lines: command history
			if line != "" {
				config.History = append(config.History, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return config, nil
}

// addToHistory adds a command to the history and maintains the maximum size
func (c *CLIConfig) addToHistory(command string) {
	// Don't add empty commands or duplicate consecutive commands
	if command == "" {
		return
	}
	if len(c.History) > 0 && c.History[len(c.History)-1] == command {
		return
	}

	c.History = append(c.History, command)

	// Keep only the last maxHistorySize entries
	if len(c.History) > maxHistorySize {
		c.History = c.History[len(c.History)-maxHistorySize:]
	}
}

// loadCurrentNodeFromConfig loads the current node from saved UUID
func loadCurrentNodeFromConfig(uuid string) (antbox.Node, error) {
	if uuid == "" || uuid == "--root--" {
		return antbox.Node{
			UUID:     "--root--",
			Title:    "root",
			Mimetype: "application/vnd.antbox.folder",
		}, nil
	}

	// Try to load the saved node
	node, err := client.GetNode(uuid)
	if err != nil {
		// If we can't load the saved node, fall back to root
		fmt.Printf("Warning: Could not restore previous location (%s), starting at root\n", uuid)
		return antbox.Node{
			UUID:     "--root--",
			Title:    "root",
			Mimetype: "application/vnd.antbox.folder",
		}, nil
	}

	return *node, nil
}

// saveCurrentState saves the current CLI state to disk
func saveCurrentState() {
	if client == nil {
		return // Not initialized yet
	}

	config := &CLIConfig{
		CurrentNodeUUID: currentNode.UUID,
		History:         cliHistory,
	}

	if err := saveConfig(config); err != nil {
		// Silently ignore save errors to avoid disrupting CLI flow
		// Could add debug logging here if needed
	}
}

// restoreFromConfig restores CLI state from saved configuration
func restoreFromConfig() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Restore current node
	restoredNode, err := loadCurrentNodeFromConfig(config.CurrentNodeUUID)
	if err != nil {
		return fmt.Errorf("failed to restore current node: %v", err)
	}

	currentNode = restoredNode

	// Restore command history
	cliHistory = config.History

	// Load current folder contents
	if nodes, err := client.ListNodes(currentNode.UUID); err == nil {
		currentNodes = nodes
	}

	return nil
}

// addCommandToHistory adds a command to the CLI history and saves state
func addCommandToHistory(command string) {
	// Skip certain commands from history
	skipCommands := map[string]bool{
		"help":    true,
		"aliases": true,
		"status":  true,
		"exit":    true,
	}

	commandParts := strings.Fields(command)
	if len(commandParts) > 0 {
		cmdName := commandParts[0]
		if !skipCommands[cmdName] {
			cliHistory = append(cliHistory, command)

			// Keep only the last maxHistorySize entries
			if len(cliHistory) > maxHistorySize {
				cliHistory = cliHistory[len(cliHistory)-maxHistorySize:]
			}

			// Save state after each command
			go saveCurrentState() // Save asynchronously to avoid blocking
		}
	}
}

// getConfigInfo returns information about the configuration file
func getConfigInfo() (string, bool, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return "", false, err
	}

	_, err = os.Stat(configPath)
	exists := !os.IsNotExist(err)

	return configPath, exists, nil
}

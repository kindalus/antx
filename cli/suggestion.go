package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

func getNodeSuggestions(word string, filter func(node antbox.Node) bool) []prompt.Suggest {
	var nodeSuggestions []prompt.Suggest
	addedUUIDs := make(map[string]bool) // Track added UUIDs to avoid duplicates

	for _, node := range currentNodes {
		if filter != nil && !filter(node) {
			continue
		}

		if (strings.HasPrefix(strings.ToLower(node.Title), strings.ToLower(word)) ||
			strings.HasPrefix(strings.ToLower(node.UUID), strings.ToLower(word))) &&
			!addedUUIDs[node.UUID] {

			var description string
			if node.Mimetype == "application/vnd.antbox.folder" {
				description = node.Title
			} else {
				description = node.Title
			}

			nodeSuggestions = append(nodeSuggestions, prompt.Suggest{
				Text:        node.UUID,
				Description: description,
			})
			addedUUIDs[node.UUID] = true
		}
	}
	return nodeSuggestions
}

func folderFilter(node antbox.Node) bool {
	return node.Mimetype == "application/vnd.antbox.folder" || node.Mimetype == "application/vnd.antbox.smartfolder"
}

func getFileSystemSuggestions(partialPath string) []prompt.Suggest {
	var suggestions []prompt.Suggest

	// If partialPath is empty, start from current directory
	if partialPath == "" {
		partialPath = "."
	}

	// Expand tilde to home directory
	if strings.HasPrefix(partialPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			if partialPath == "~" {
				partialPath = homeDir
			} else if strings.HasPrefix(partialPath, "~/") {
				partialPath = strings.Replace(partialPath, "~", homeDir, 1)
			}
		}
	}

	// Get the directory and filename parts
	dir := filepath.Dir(partialPath)
	filename := filepath.Base(partialPath)

	// If the path ends with a separator, we're looking in that directory
	if strings.HasSuffix(partialPath, string(filepath.Separator)) {
		dir = partialPath
		filename = ""
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless explicitly requested
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(filename, ".") {
			continue
		}

		// Filter by filename prefix
		if filename != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(filename)) {
			continue
		}

		// Build the full path
		var fullPath string
		if dir == "." {
			fullPath = name
		} else {
			fullPath = filepath.Join(dir, name)
		}

		// Add appropriate description and suffix
		var description string
		if entry.IsDir() {
			description = "📁 Directory"
			fullPath += string(filepath.Separator)
		} else {
			info, err := entry.Info()
			if err == nil {
				size := info.Size()
				description = fmt.Sprintf("📄 File (%s)", formatFileSize(size))
			} else {
				description = "📄 File"
			}
		}

		text := fullPath
		if strings.Contains(fullPath, " ") {
			text = `"` + fullPath + `"`
		}
		suggestions = append(suggestions, prompt.Suggest{
			Text:        text,
			Description: description,
		})
	}

	return suggestions
}

func formatFileSize(size int64) string {
	if size == 0 {
		return "0 B"
	}

	const unit = 1024
	units := []string{"B", "K", "M", "G", "T", "P"}

	sizeFloat := float64(size)
	unitIndex := 0

	for sizeFloat >= unit && unitIndex < len(units)-1 {
		sizeFloat /= unit
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", sizeFloat, units[unitIndex])
	}

	return fmt.Sprintf("%.1f %s", sizeFloat, units[unitIndex])
}

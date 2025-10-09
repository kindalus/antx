package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	markdown "go.xrstf.de/go-term-markdown"
)

type DocsCommand struct{}

func (c *DocsCommand) GetName() string {
	return "docs"
}

func (c *DocsCommand) GetDescription() string {
	return "List all available documents or display a specific document"
}

func (c *DocsCommand) Execute(args []string) {
	if len(args) == 0 {
		// List all documents
		docs, err := client.ListDocs()
		if err != nil {
			fmt.Println("Error listing documents:", err)
			return
		}

		if len(docs) == 0 {
			fmt.Println("No documents available.")
			return
		}

		fmt.Println("Available documents:")
		fmt.Println()
		for _, doc := range docs {
			fmt.Printf("UUID: %s\n", doc.UUID)
			fmt.Printf("  Description: %s\n", doc.Description)
			fmt.Println()
		}
		return
	}

	// Display specific document
	docUUID := args[0]

	// Get document content
	docContent, err := client.GetDoc(docUUID)
	if err != nil {
		fmt.Println("Error getting document:", err)
		return
	}

	// Render markdown content
	result := markdown.Render(docContent, 100, 11)
	fmt.Print(string(result))
}

func (c *DocsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) <= 2 {
		// First argument - suggest document UUIDs
		word := d.GetWordBeforeCursor()

		// Get cached docs or fetch them
		docs, err := client.ListDocs()
		if err != nil {
			return []prompt.Suggest{}
		}

		var suggests []prompt.Suggest
		for _, doc := range docs {
			if strings.HasPrefix(strings.ToLower(doc.UUID), strings.ToLower(word)) ||
				strings.HasPrefix(strings.ToLower(doc.Description), strings.ToLower(word)) {
				suggests = append(suggests, prompt.Suggest{
					Text:        doc.UUID,
					Description: doc.Description,
				})
			}
		}
		return suggests
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&DocsCommand{})
}

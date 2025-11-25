package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type WhoAmICommand struct{}

func (c *WhoAmICommand) GetName() string {
	return "whoami"
}

func (c *WhoAmICommand) GetDescription() string {
	return "Show the current authenticated user"
}

func (c *WhoAmICommand) Execute(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: whoami")
		return
	}

	user, err := client.GetCurrentUser()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Current user:")
	fmt.Printf("  Email : %s\n", user.Email)
	if user.Name != "" {
		fmt.Printf("  Name  : %s\n", user.Name)
	}
	if len(user.Groups) > 0 {
		fmt.Printf("  Groups: %s\n", strings.Join(user.Groups, ", "))
	}
}

func (c *WhoAmICommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&WhoAmICommand{})
}

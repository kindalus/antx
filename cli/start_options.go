package cli

import (
	"fmt"
	"strings"
)

const defaultLightrayClientID = "terminal-cli"

func DefaultLightrayClientID() string {
	return defaultLightrayClientID
}

// StartOptions configures the interactive antx shell startup.
type StartOptions struct {
	ServerURL        string
	APIKey           string
	Root             string
	JWT              string
	Debug            bool
	Lightray         bool
	LightrayClientID string
}

func (o StartOptions) Validate() error {
	if strings.TrimSpace(o.ServerURL) == "" {
		return fmt.Errorf("server URL is required")
	}

	if !o.Lightray {
		return nil
	}

	var conflicts []string
	if o.APIKey != "" {
		conflicts = append(conflicts, "--api-key")
	}
	if o.Root != "" {
		conflicts = append(conflicts, "--root")
	}
	if o.JWT != "" {
		conflicts = append(conflicts, "--jwt")
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("--lightray cannot be used with %s", strings.Join(conflicts, ", "))
	}

	return nil
}

func (o StartOptions) effectiveLightrayClientID() string {
	clientID := strings.TrimSpace(o.LightrayClientID)
	if clientID == "" {
		return defaultLightrayClientID
	}
	return clientID
}

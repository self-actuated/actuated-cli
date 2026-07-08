package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	PlatformGitHub = "github"
	PlatformGitLab = "gitlab"
)

const basePath = "$HOME/.actuated"

// getPlatform returns the platform for the current controller.
// It first checks the config file for a controller entry matching the
// ACTUATED_URL. If not found, it falls back to the legacy $HOME/.actuated/PLATFORM file.
// Defaults to "github" if neither source has a value.
func getPlatform() string {
	cc, _, found, err := getControllerConfig()
	if err == nil && found && cc.Platform != "" {
		return cc.Platform
	}

	// Fallback: read legacy PLATFORM file
	platformFile := os.ExpandEnv(path.Join(basePath, "PLATFORM"))

	data, err := os.ReadFile(platformFile)
	if err != nil {
		return PlatformGitHub
	}

	platform := strings.TrimSpace(string(data))
	if platform == PlatformGitLab {
		return PlatformGitLab
	}

	return PlatformGitHub
}

// validatePlatform checks that the given platform string is valid.
func validatePlatform(platform string) error {
	switch platform {
	case PlatformGitHub, PlatformGitLab:
		return nil
	default:
		return fmt.Errorf("unsupported platform: %q, supported values: %s, %s", platform, PlatformGitHub, PlatformGitLab)
	}
}

// checkGitHubOnly returns an error if the current platform is not GitHub.
// Use this at the start of command handlers that have not been implemented
// for other platforms yet.
func checkGitHubOnly(commandName string) error {
	if p := getPlatform(); p != PlatformGitHub {
		return fmt.Errorf("the %q command is not supported for platform %q", commandName, p)
	}
	return nil
}

package wguctl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GetPublicKeyFromConfig extracts the public key from a WireGuard config file using wgu
func GetPublicKeyFromConfig(ctx context.Context, config Config) (string, error) {
	wguCmd := exec.CommandContext(ctx, config.GetExePath(), "pubkeyconf", config.ConfigPath)

	var stderr bytes.Buffer
	wguCmd.Stderr = &stderr

	output, err := wguCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute '%s' - %w - stderr: '%s'",
			wguCmd.String(), err, strings.TrimSpace(stderr.String()))
	}

	// Trim whitespace from output
	pubkey := strings.TrimSpace(string(output))
	return pubkey, nil
}

// CreateConfig makes a default config file
func CreateConfig(ctx context.Context, config Config, profileName string) error {
	wguCmd := exec.CommandContext(ctx, config.ExePath, "genconf", "-n", profileName+".conf", config.ConfigPath)

	var stderr bytes.Buffer
	wguCmd.Stderr = &stderr

	_, err := wguCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute '%s' - %w - stderr: '%s'",
			wguCmd.String(), err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

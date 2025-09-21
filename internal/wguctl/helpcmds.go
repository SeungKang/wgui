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
			wguCmd.String(), err, stderr.String())
	}

	// Trim whitespace from output
	pubkey := strings.TrimSpace(string(output))
	return pubkey, nil
}

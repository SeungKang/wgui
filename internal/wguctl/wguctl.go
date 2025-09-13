package wguctl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type Wgu struct {
	process *exec.Cmd
	stderr chan string
}

type WguConfig struct {
	ConfigPath string
}

func StartWgu(ctx context.Context, config WguConfig) (*Wgu, error) {
	wgu := exec.CommandContext(ctx, "wgu", "up", config.ConfigPath)

	stdout, err := wgu.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe - %w", err)
	}

	stderr, err := wgu.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe - %w", err)
	}

	err = wgu.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start wgu - %w", err)
	}

	stderrLines := make(chan string)

	go func() {
		stderrScanner := bufio.NewScanner(stderr)

		for stderrScanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case stderrLines <- stderrScanner.Text():
				// keep going
			}
		}
	}()

	startResult := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)

		if !scanner.Scan() {
			if wgu.ProcessState.Exited() {
				startResult <-fmt.Errorf("failed to start wgu - command exited - %d", wgu.ProcessState.ExitCode())
				return
			}

			startResult <-fmt.Errorf("failed to scan stdout - %w", scanner.Err())
			return
		}

		if scanner.Text() != "ready" {
			startResult <-fmt.Errorf("failed to read 'ready' string from stdout - got: %q", scanner.Text())
			return
		}

		startResult <- nil
	}()

	timeout := time.After(time.Second)

	select {
	case <-timeout:
		_ = wgu.Process.Kill()
		return nil, errors.New("timed out waiting for wgu to become ready")
	case err:=<-startResult:
		if err != nil {
			_ = wgu.Process.Kill()
			return nil, fmt.Errorf("failed to get 'ready' result - %w", err)
		}

		return &Wgu{process: wgu, stderr: stderrLines}, nil
	}
}

func (w *Wgu) Stop() error {
	return w.process.Process.Kill()
}

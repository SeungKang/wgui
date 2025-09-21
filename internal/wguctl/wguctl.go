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
	stderr  chan string
}

type Config struct {
	ExePath    string
	ConfigPath string
}

func (o *Config) GetExePath() string {
	if o.ExePath == "" {
		return "wgu"
	}

	return o.ExePath
}

func StartWgu(ctx context.Context, config Config) (*Wgu, error) {
	wgu := exec.CommandContext(ctx, "wgu", "up", "-c", config.ConfigPath)

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

	// TODO we need to store this channel in the struct
	exited := make(chan error, 1)

	go func() {
		exited <- wgu.Wait()
	}()

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

	isReady := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)

		if !scanner.Scan() {
			isReady <- fmt.Errorf("failed to scan stdout - %w", scanner.Err())
			return
		}

		if scanner.Text() != "ready" {
			isReady <- fmt.Errorf("failed to read 'ready' string from stdout - got: %q", scanner.Text())
			return
		}

		isReady <- nil
	}()

	timeout := time.After(3 * time.Second)

	select {
	case <-timeout:
		_ = wgu.Process.Kill()
		return nil, errors.New("timed out waiting for wgu to become ready")
	case err := <-exited:
		if err != nil {
			return nil, fmt.Errorf("wgu process exited unexpectedly while waiting for 'ready' - %w", err)
		}

		return nil, fmt.Errorf("wgu process exited unexpectedly without error while waiting for 'ready'")
	case err := <-isReady:
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

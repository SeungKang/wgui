package wguctl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type Wgu struct {
	once    sync.Once
	process *exec.Cmd
	stdin   io.WriteCloser
}

type Config struct {
	ExePath    string
	ConfigPath string
	OptStderr  chan<- string
}

func (o *Config) GetExePath() string {
	if o.ExePath == "" {
		return "wgu"
	}

	return o.ExePath
}

func StartWgu(ctx context.Context, config Config) (*Wgu, error) {
	wgu := exec.CommandContext(ctx, config.ExePath, "up", "-c", config.ConfigPath) // TODO should this be config.GetExePath()?

	stdin, err := wgu.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe - %w", err)
	}

	stdout, err := wgu.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe - %w", err)
	}

	var stderr io.Reader
	if config.OptStderr != nil {
		stderr, err = wgu.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe - %w", err)
		}
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

	if stderr != nil {
		go func() {
			stderrScanner := bufio.NewScanner(stderr)

			for stderrScanner.Scan() {
				select {
				case <-ctx.Done():
					return
				case config.OptStderr <- stderrScanner.Text():
					// keep going
				}
			}
		}()
	}

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

		return &Wgu{process: wgu, stdin: stdin}, nil
	}
}

func (o *Wgu) Stop() error {
	o.stdin.Close()
	return o.process.Process.Kill()
}

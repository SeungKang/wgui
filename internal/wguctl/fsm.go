package wguctl

import (
	"context"
	"sync"
)

type FsmState int

const (
	UnknownFsmState FsmState = iota
	DisconnectedFsmState
	ConnectedFsmState
	ErrorFsmState
	DisconnectingFsmState
	ConnectingFsmState
)

func NewFsm(ctx context.Context) *Fsm {
	fsm := &Fsm{
		events:   make(chan interface{}, 10),
		state:    DisconnectedFsmState,
		stderrCh: make(chan string),
		done:     make(chan struct{}),
	}

	go fsm.loop(ctx)

	go fsm.handleStderr(ctx)

	return fsm
}

type Fsm struct {
	wgu        *Wgu
	events     chan interface{}
	rwMutex    sync.RWMutex
	state      FsmState
	lastError  error
	stderrRWMu sync.RWMutex
	stderr     string
	stderrCh   chan string
	done       chan struct{}
}

func (o *Fsm) Connect(ctx context.Context, config Config) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case o.events <- connectFsmEvent{config: config}:
		return nil
	}
}

func (o *Fsm) Disconnect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case o.events <- disconnectFsmEvent{}:
		return nil
	}
}

type connectFsmEvent struct {
	config Config
}

type disconnectFsmEvent struct {
}

func (o *Fsm) loop(ctx context.Context) {
	defer close(o.done)

	for {
		select {
		case <-ctx.Done():
			if o.wgu != nil {
				_ = o.disconnect(ctx)
			}
			return
		case e := <-o.events:
			o.processEvent(ctx, e)
		}
	}
}

func (o *Fsm) Done() <-chan struct{} {
	return o.done
}

func (o *Fsm) processEvent(ctx context.Context, event interface{}) {
	switch e := event.(type) {
	case connectFsmEvent:
		o.rwMutex.Lock()
		o.state = ConnectingFsmState
		o.rwMutex.Unlock()

		err := o.connect(ctx, e.config)

		o.rwMutex.Lock()
		defer o.rwMutex.Unlock()

		if err != nil {
			o.state = ErrorFsmState
			o.lastError = err
		} else {
			o.state = ConnectedFsmState
			o.lastError = nil
		}
	case disconnectFsmEvent:
		o.rwMutex.Lock()
		o.state = DisconnectingFsmState
		o.rwMutex.Unlock()

		err := o.disconnect(ctx)

		o.rwMutex.Lock()
		defer o.rwMutex.Unlock()

		if err != nil {
			o.state = ErrorFsmState
			o.lastError = err
		} else {
			o.state = DisconnectedFsmState
			o.lastError = nil
		}
	}
}

func (o *Fsm) connect(ctx context.Context, config Config) error {
	if o.wgu != nil {
		_ = o.wgu.Stop()
		o.wgu = nil
	}

	config.OptStderr = o.stderrCh

	wgu, err := StartWgu(ctx, config)
	if err != nil {
		return err
	}

	o.wgu = wgu
	return nil
}

func (o *Fsm) disconnect(ctx context.Context) error {
	_ = o.wgu.Stop()
	o.wgu = nil
	return nil
}

func (o *Fsm) State() (FsmState, error) {
	o.rwMutex.RLock()
	defer o.rwMutex.RUnlock()

	return o.state, o.lastError
}

func (o *Fsm) Stderr() string {
	o.stderrRWMu.RLock()
	defer o.stderrRWMu.RUnlock()

	return o.stderr
}

func (o *Fsm) handleStderr(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case line := <-o.stderrCh:
			o.stderrRWMu.Lock()

			o.stderr += line + "\n"

			o.stderrRWMu.Unlock()
		}
	}
}

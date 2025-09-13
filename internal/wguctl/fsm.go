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
)

func NewFsm(ctx context.Context) *Fsm {
	fsm := &Fsm{
		events: make(chan interface{}, 10),
		state:  DisconnectedFsmState,
	}

	go fsm.loop(ctx)

	return fsm
}

type Fsm struct {
	wgu       *Wgu
	events    chan interface{}
	rwMutex   sync.RWMutex
	state     FsmState
	lastError error
}

type connectFsmEvent struct {
	config WguConfig
}

type disconnectFsmEvent struct {
}

func (o *Fsm) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-o.events:
			o.processEvent(ctx, e)
		}
	}
}

func (o *Fsm) processEvent(ctx context.Context, event interface{}) {
	switch e := event.(type) {
	case connectFsmEvent:
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

func (o *Fsm) connect(ctx context.Context, config WguConfig) error {
	if o.wgu != nil {
		_ = o.wgu.Stop()
		o.wgu = nil
	}

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

package main

import (
	"context"
	"os"
	"path/filepath"
	"wugui/internal/wguctl"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type State struct {
	wgu           *wguctl.Fsm
	connected     bool
	connectButton *widget.Clickable
	list          *widget.List
	theme         *material.Theme
	win           *app.Window
	configDirPath string

	// new config
	profileNameEditor *widget.Editor
	configEditor      *widget.Editor

	// sidebar
	sidebarList     *widget.List
	profiles        []string
	profileClicks   []widget.Clickable
	selectedProfile int
}

func NewState(ctx context.Context, w *app.Window) *State {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	s := &State{
		wgu:           wguctl.NewFsm(ctx),
		configEditor:  new(widget.Editor),
		connectButton: new(widget.Clickable),
		list: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		theme:             material.NewTheme(),
		win:               w,
		sidebarList:       &widget.List{List: layout.List{Axis: layout.Vertical}},
		profiles:          []string{"+"},
		profileClicks:     make([]widget.Clickable, 1),
		profileNameEditor: new(widget.Editor),
		configDirPath:     filepath.Join(homeDir, ".wgu"),
	}

	s.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	return s
}

func (s *State) Run(ctx context.Context, w *app.Window) error {
	events := make(chan event.Event)
	acks := make(chan struct{})

	go func() {
		for {
			ev := w.Event()
			events <- ev
			<-acks
			_, ok := ev.(app.DestroyEvent)
			if ok {
				return
			}
		}
	}()

	var ops op.Ops
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-events:
			switch e := e.(type) {
			case app.DestroyEvent:
				acks <- struct{}{}
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				if s.profiles[s.selectedProfile] == "+" {
					s.renderNewProfileFrame(ctx, gtx)
				} else {
					s.renderProfileFrame(ctx, gtx)
				}
				e.Frame(gtx.Ops)
			}

			acks <- struct{}{}
		}
	}
}

func (s *State) AddProfile(name string) {
	s.profiles = append(s.profiles, name)
	s.profileClicks = append(s.profileClicks, widget.Clickable{})
}

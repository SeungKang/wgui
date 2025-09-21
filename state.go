package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	wguDir        string
	wguExePath    string
	errLogger     *log.Logger

	// new config
	profileNameEditor *widget.Editor
	configEditor      *widget.Editor

	profiles *profileState
}

type profileState struct {
	profileList     *widget.List
	profiles        []profileConfig
	profileClicks   []widget.Clickable
	selectedProfile int
}

type profileConfig struct {
	name       string
	configPath string
	pubkey     string
}

func (o *profileConfig) refresh(ctx context.Context, wguExePath string) error {
	pubkey, err := wguctl.GetPublicKeyFromConfig(ctx, wguctl.Config{
		ExePath:    wguExePath,
		ConfigPath: o.configPath,
	})
	if err != nil {
		return err
	}

	o.pubkey = pubkey

	return nil
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
		theme: material.NewTheme(),
		win:   w,
		profiles: &profileState{
			profileList:     &widget.List{List: layout.List{Axis: layout.Vertical}},
			profileClicks:   []widget.Clickable{},
			selectedProfile: 0,
		},
		profileNameEditor: new(widget.Editor),
		wguDir:            filepath.Join(homeDir, ".wgu"),
		wguExePath:        "wgu",
		errLogger:         log.Default(),
	}

	s.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	err = s.loadProfiles()
	if err != nil {
		panic(err)
	}

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
				if s.profiles.profiles[s.profiles.selectedProfile].name == "+" {
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

func (s *State) loadProfiles() error {
	// Ensure .wgu directory exists
	err := os.MkdirAll(s.wguDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create directory %s - %w", s.wguDir, err)
	}

	// Read all .conf paths from the directory
	paths, err := filepath.Glob(filepath.Join(s.wguDir, "*.conf"))
	if err != nil {
		return fmt.Errorf("failed to get all .conf paths - %v", err)
	}

	// Extract profile names and sort them consistently
	var profileConfigs []profileConfig

	for _, path := range paths {
		baseName := filepath.Base(path)
		profileName := strings.TrimSuffix(baseName, ".conf")
		profileConfigs = append(profileConfigs, profileConfig{
			name:       profileName,
			configPath: path,
		})
	}

	// Ensure consistent ordering
	sort.Slice(profileConfigs, func(i, j int) bool {
		return profileConfigs[i].name < profileConfigs[j].name
	})

	// Initialize profiles with sorted profiles first, then "+" button at the end
	s.profiles.profiles = make([]profileConfig, 0, len(profileConfigs)+1)
	s.profiles.profiles = append(s.profiles.profiles, profileConfigs...)
	s.profiles.profiles = append(s.profiles.profiles, profileConfig{
		name:       "+",
		configPath: "",
	})

	// Initialize clickable widgets for all profiles
	s.profiles.profileClicks = make([]widget.Clickable, len(s.profiles.profiles))

	// Set initial selection - if we have profiles, select the first one; otherwise select "+"
	s.profiles.selectedProfile = 0

	return nil
}

func (s *State) RefreshProfiles() {
	err := s.loadProfiles()
	if err != nil {
		// Handle error appropriately - could log or show in UI
		fmt.Printf("Error refreshing profiles: %v\n", err)
	}
	if s.win != nil {
		s.win.Invalidate()
	}
}

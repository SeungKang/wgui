package main

import (
	"context"
	"fmt"
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

	// new config
	profileNameEditor *widget.Editor
	configEditor      *widget.Editor

	// sidebar
	sidebarList     *widget.List
	profiles        []string
	profilePaths    map[string]string // maps profile names to file paths
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
		profiles:          []string{},           // Initialize empty, will be populated by loadProfiles
		profileClicks:     []widget.Clickable{}, // Initialize empty
		profileNameEditor: new(widget.Editor),
		profilePaths:      make(map[string]string), // Initialize the map!
		wguDir:            filepath.Join(homeDir, ".wgu"),
		selectedProfile:   0, // Will be adjusted after loading profiles
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

func (s *State) loadProfiles() error {
	// Ensure .wgu directory exists
	err := os.MkdirAll(s.wguDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create directory %s - %w", s.wguDir, err)
	}

	// Read all .conf files from the directory
	files, err := filepath.Glob(filepath.Join(s.wguDir, "*.conf"))
	if err != nil {
		return fmt.Errorf("failed to get all .conf files - %v", err)
	}

	// Extract profile names and sort them consistently
	var profileNames []string
	profilePaths := make(map[string]string)

	for _, file := range files {
		baseName := filepath.Base(file)
		profileName := strings.TrimSuffix(baseName, ".conf")
		profileNames = append(profileNames, profileName)
		profilePaths[profileName] = file
	}

	// Ensure consistent ordering
	sort.Strings(profileNames)

	// Initialize profiles with sorted profiles first, then "+" button at the end
	s.profiles = make([]string, 0, len(profileNames)+1)
	s.profiles = append(s.profiles, profileNames...)
	s.profiles = append(s.profiles, "+")

	// Update the profilePaths map
	s.profilePaths = profilePaths

	// Initialize clickable widgets for all profiles
	s.profileClicks = make([]widget.Clickable, len(s.profiles))

	// Set initial selection - if we have profiles, select the first one; otherwise select "+"
	s.selectedProfile = 0

	return nil
}

func (s *State) GetProfilePath(profileName string) string {
	if path, exists := s.profilePaths[profileName]; exists {
		return path
	}
	return ""
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

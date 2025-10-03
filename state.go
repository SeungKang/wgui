package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	// sidebar
	newProfileButton    *widget.Clickable
	refreshIconButton   *widget.Clickable
	sidebarProfilesList *widget.List

	// profile_frame
	pubkeySelectable  *widget.Selectable
	copyIconButton    *widget.Clickable
	copiedMessageTime time.Time
	connectButton     *widget.Clickable
	editButton        *widget.Clickable
	logsList          *widget.List
	logSelectables    *widget.Selectable

	// new_profile_frame
	profileNameEditor *widget.Editor
	configEditor      *widget.Editor
	saveButton        *widget.Clickable
	cancelButton      *widget.Clickable
	deleteButton      *widget.Clickable

	// window
	theme *material.Theme
	win   *app.Window

	wguDir        string
	wguExePath    string
	errLogger     *log.Logger
	currentUiMode uiMode
	profiles      *profileState
}

type uiMode int

const (
	newProfileUiMode uiMode = iota
	editProfileUiMode
	viewProfileUiMode
)

type profileState struct {
	profileList   *widget.List
	profiles      []profileConfig
	profileClicks []widget.Clickable
	selectedIndex int
	events        chan profileEvent
}

func (o *profileState) selected() *profileConfig {
	return &o.profiles[o.selectedIndex]
}

type profileEvent struct {
	name string
}

type profileConfig struct {
	name           string
	configPath     string
	pubkey         string
	lastReadConfig string
	wgu            *wguctl.Fsm
	lastErrMsg     string
}

func (o *profileConfig) refresh(ctx context.Context, wguExePath string, logger *log.Logger) {
	err := o.refreshWithErr(ctx, wguExePath)
	if err != nil {
		o.lastErrMsg = err.Error()
		logger.Printf("failed to refresh profile - %v", err)
	} else {
		o.lastErrMsg = ""
	}
}

func (o *profileConfig) refreshWithErr(ctx context.Context, wguExePath string) error {
	config, err := os.ReadFile(o.configPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s - %v", o.configPath, err)
	}

	o.lastReadConfig = string(config)

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

	wguPath, err := wguExePath()
	if err != nil {
		panic(err)
	}

	s := &State{
		newProfileButton:  new(widget.Clickable),
		refreshIconButton: new(widget.Clickable),
		sidebarProfilesList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		copyIconButton: new(widget.Clickable),
		connectButton:  new(widget.Clickable),
		editButton:     new(widget.Clickable),
		logsList: &widget.List{
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: true,
			},
		},
		logSelectables:    new(widget.Selectable),
		profileNameEditor: new(widget.Editor),
		configEditor:      new(widget.Editor),
		saveButton:        new(widget.Clickable),
		cancelButton:      new(widget.Clickable),
		deleteButton:      new(widget.Clickable),
		theme:             material.NewTheme(),
		win:               w,
		profiles: &profileState{
			profileList: &widget.List{List: layout.List{Axis: layout.Vertical}},
			events:      make(chan profileEvent),
		},
		wguDir:        filepath.Join(homeDir, ".wgu"),
		wguExePath:    wguPath,
		errLogger:     log.Default(),
		currentUiMode: newProfileUiMode,
	}

	s.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	err = s.loadProfiles(ctx)
	if err != nil {
		panic(err)
	}

	if len(s.profiles.profiles) > 0 {
		s.currentUiMode = viewProfileUiMode
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
				switch s.currentUiMode {
				case newProfileUiMode:
					s.renderNewProfileFrame(ctx, gtx)
				case editProfileUiMode:
					s.renderNewProfileFrame(ctx, gtx)
				case viewProfileUiMode:
					s.renderProfileFrame(ctx, gtx)
				}

				e.Frame(gtx.Ops)
			}

			acks <- struct{}{}
		case e := <-s.profiles.events:
			if e.name == s.profiles.selected().name {
				s.win.Invalidate()
			}
		}
	}
}

func (s *State) loadProfiles(ctx context.Context) error {
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
		config := profileConfig{
			name:       profileName,
			configPath: path,
			wgu: wguctl.NewFsm(ctx, wguctl.FsmConfig{
				OnNewStderr: func(ctx context.Context) {
					select {
					case <-ctx.Done():
					case s.profiles.events <- profileEvent{name: profileName}:
					}
				},
			}),
		}

		config.refresh(ctx, s.wguExePath, s.errLogger)

		profileConfigs = append(profileConfigs, config)
	}

	// Ensure consistent ordering
	sort.Slice(profileConfigs, func(i, j int) bool {
		return profileConfigs[i].name < profileConfigs[j].name
	})

	// Initialize profiles with sorted profiles
	s.profiles.profiles = profileConfigs

	// Initialize clickable widgets for all profiles
	s.profiles.profileClicks = make([]widget.Clickable, len(s.profiles.profiles))

	return nil
}

func (s *State) RefreshProfiles(ctx context.Context) error {
	err := s.loadProfiles(ctx)
	if err != nil {
		s.errLogger.Printf("Error refreshing profiles: %v", err)
	}

	if s.win != nil {
		s.win.Invalidate()
	}

	return nil
}

package main

import (
	"context"
	"os"
	"path/filepath"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func (s *State) renderNewProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor)

	widgets := []layout.Widget{
		func(gtx C) D { return s.renderSpacer(gtx, unit.Dp(16)) },
		s.formField("Name", s.profileNameEditor, unit.Dp(80)),
		s.formField("Config", s.configEditor, unit.Dp(200)),

		func(gtx C) D {
			return s.renderButton(ctx, gtx, "Save", PurpleColor, s.saveButton, func() { s.saveNewProfile(ctx) })
		},

		func(gtx C) D { return s.renderErrorMessage(gtx, "this is an error message") },
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// LEFT: Sidebar
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(ctx, gtx)
		}),

		// RIGHT: Main content
		layout.Flexed(1, func(gtx C) D {
			return material.List(s.theme, s.list).Layout(gtx, len(widgets), func(gtx C, i int) D {
				return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
			})
		}),
	)
}

// Replaces the separate label/editor widgets with a grouped field
func (s *State) formField(label string, ed *widget.Editor, h unit.Dp) layout.Widget {
	return func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Label (left-aligned) with a small gap to the editor
			layout.Rigid(func(gtx C) D {
				l := material.Label(s.theme, 16, label)
				l.Color = WhiteColor
				l.Alignment = text.Start // left align
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, l.Layout)
			}),
			// Editor
			layout.Rigid(func(gtx C) D {
				return s.renderTextEditor(gtx, ed, "", h)
			}),
		)
	}
}

func (s *State) saveNewProfile(ctx context.Context) {
	profileName := s.profileNameEditor.Text()
	configContent := s.configEditor.Text()

	if profileName == "" || configContent == "" {
		s.errLogger.Printf("Profile name or config content is empty")
		return
	}

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		s.errLogger.Printf("Error getting home directory: %v", err)
		return
	}

	wguDir := filepath.Join(homeDir, ".wgu")

	// Create .wgu directory if it doesn't exist
	if err := os.MkdirAll(wguDir, 0755); err != nil {
		s.errLogger.Printf("Error creating .wgu directory: %v", err)
		return
	}

	// Create the config file path
	configPath := filepath.Join(wguDir, profileName+".conf")

	// Write the config to file
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		s.errLogger.Printf("Error saving config file: %v", err)
		return
	}

	// Clear the editors
	s.profileNameEditor.SetText("")
	s.configEditor.SetText("")

	// Refresh the profiles list
	s.RefreshProfiles(ctx)

	// Switch to the newly created profile (it will be in the sorted list, not at the end)
	for i, profile := range s.profiles.profiles {
		if profile.name == profileName {
			s.profiles.selectedProfile = i
			s.currentUiMode = viewProfileUiMode
			break
		}
	}
}
func (s *State) handleNewProfileEditorUpdates(gtx layout.Context) {
	// Profile name editor updates
	for {
		updateEvent, ok := s.profileNameEditor.Update(gtx)
		if !ok {
			break
		}

		if _, ok := updateEvent.(widget.ChangeEvent); ok {
			var err error
			if err != nil {
				s.errLogger.Printf("failed to render profile name editor - %v", err)
			}
		}
	}

	// Config editor updates
	for {
		updateEvent, ok := s.configEditor.Update(gtx)
		if !ok {
			break
		}

		if _, ok := updateEvent.(widget.ChangeEvent); ok {
			var err error
			if err != nil {
				s.errLogger.Printf("failed to render config editor - %v", err)
			}
		}
	}
}

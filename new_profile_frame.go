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

// renderNewProfileFrame is the main layout for creating/editing a profile
func (s *State) renderNewProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor)

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(ctx, gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return s.renderNewProfileContent(ctx, gtx)
		}),
	)
}

// renderNewProfileContent contains the form and action bar
func (s *State) renderNewProfileContent(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return s.renderProfileForm(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return s.renderFormActionBar(ctx, gtx)
		}),
	)
}

// renderProfileForm displays the scrollable form with name and config fields
func (s *State) renderProfileForm(gtx layout.Context) layout.Dimensions {
	form := []layout.Widget{
		func(gtx C) D { return s.renderSpacer(gtx, unit.Dp(16)) },
		s.formField("Name", s.profileNameEditor, unit.Dp(30)),
		s.formField("Config", s.configEditor, unit.Dp(300)),
	}

	s.handleProfileEditorUpdates(gtx)

	return material.List(s.theme, s.sidebarProfilesList).Layout(gtx, len(form), func(gtx C, i int) D {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, form[i])
	})
}

// renderFormActionBar shows Save/Cancel buttons and error messages
func (s *State) renderFormActionBar(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top: unit.Dp(8), Left: unit.Dp(16),
		Right: unit.Dp(16), Bottom: unit.Dp(16),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return s.renderFormButtons(ctx, gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return s.renderFormErrorSection(gtx)
			}),
		)
	})
}

// renderFormButtons displays the Save and Cancel buttons
func (s *State) renderFormButtons(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return s.renderSaveButton(ctx, gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return s.renderCancelButton(ctx, gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if s.currentUiMode == newProfileUiMode {
				return D{}
			} else {
				return s.renderDeleteButton(ctx, gtx)
			}
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Spacer{}.Layout(gtx)
		}),
	)
}

// renderDeleteButton shows the delete profile button
func (s *State) renderDeleteButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	onClick := func() {
		s.deleteProfile(ctx)

		if len(s.profiles.profiles) == 0 {
			s.currentUiMode = newProfileUiMode
			s.profileNameEditor.SetText("")
			s.configEditor.SetText("")
		} else {
			s.profiles.selectedIndex = 0
			s.currentUiMode = viewProfileUiMode
		}
	}

	return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
		return s.renderButton(gtx, "Delete", RedColor, s.deleteButton, onClick)
	})
}

// renderSaveButton shows the save button
func (s *State) renderSaveButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	onClick := func() {
		s.saveProfile(ctx)
	}
	return s.renderButton(gtx, "Save", PurpleColor, s.saveButton, onClick)
}

// renderCancelButton shows the cancel button
func (s *State) renderCancelButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	onClick := func() {
		if len(s.profiles.profiles) != 0 {
			s.currentUiMode = viewProfileUiMode
		}
	}

	return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
		return s.renderButton(gtx, "Cancel", GreyColor, s.cancelButton, onClick)
	})
}

// renderFormErrorSection displays error messages
func (s *State) renderFormErrorSection(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
		return s.renderErrorMessage(gtx, s.errLabel)
	})
}

// formField creates a labeled form field with an editor
func (s *State) formField(label string, ed *widget.Editor, h unit.Dp) layout.Widget {
	return func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return s.renderFieldLabel(gtx, label)
			}),
			layout.Rigid(func(gtx C) D {
				return s.renderTextEditor(gtx, ed, "", h)
			}),
		)
	}
}

// renderFieldLabel displays a form field label
func (s *State) renderFieldLabel(gtx layout.Context, label string) layout.Dimensions {
	l := material.Label(s.theme, 16, label)
	l.Color = WhiteColor
	l.Alignment = text.Start
	return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, l.Layout)
}

// handleProfileEditorUpdates processes editor events
func (s *State) handleProfileEditorUpdates(gtx layout.Context) {
	s.handleEditorUpdates(s.profileNameEditor, gtx)
	s.handleEditorUpdates(s.configEditor, gtx)
}

// handleEditorUpdates consumes all pending updates for an editor
func (s *State) handleEditorUpdates(ed *widget.Editor, gtx layout.Context) {
	for {
		_, ok := ed.Update(gtx)
		if !ok {
			break
		}
	}
}

// saveProfile validates and saves the profile configuration
func (s *State) saveProfile(ctx context.Context) {
	profileName := s.profileNameEditor.Text()
	configContent := s.configEditor.Text()

	if !s.validateProfileInput(profileName, configContent) {
		return
	}

	if err := s.ensureWguDirectory(); err != nil {
		return
	}

	configPath := filepath.Join(s.wguConfDir, profileName+".conf")
	if err := s.writeConfigFile(configPath, configContent); err != nil {
		return
	}

	s.clearEditors()
	s.refreshAndSelectProfile(ctx, profileName)
}

// validateProfileInput checks if profile name and config are valid
func (s *State) validateProfileInput(name, config string) bool {
	if name == "" && config == "" {
		s.errLabel = "Please enter a profile name and config file"
		s.errLogger.Printf("Profile name and config content are empty")
		return false
	}

	if name == "" {
		s.errLabel = "Please enter a profile name"
		s.errLogger.Printf("Profile name is empty")
		return false
	}

	if config == "" {
		s.errLabel = "Please enter a config file"
		s.errLogger.Printf("Config content is empty")
		return false
	}

	return true
}

// ensureWguDirectory creates the .wgu directory if it doesn't exist
func (s *State) ensureWguDirectory() error {
	if err := os.MkdirAll(s.wguConfDir, 0755); err != nil {
		s.errLogger.Printf("Error creating .wgu directory: %v", err)
		return err
	}
	return nil
}

// writeConfigFile saves the config content to a file
func (s *State) writeConfigFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		s.errLogger.Printf("Error saving config file: %v", err)
		return err
	}
	return nil
}

// clearEditors resets both editors to empty
func (s *State) clearEditors() {
	s.profileNameEditor.SetText("")
	s.configEditor.SetText("")
}

// refreshAndSelectProfile refreshes profiles and switches to the new one
func (s *State) refreshAndSelectProfile(ctx context.Context, profileName string) {
	if err := s.RefreshProfiles(ctx); err != nil {
		s.errLogger.Printf("failed to refresh profile - %v", err)
		return
	}

	for i, profile := range s.profiles.profiles {
		if profile.name == profileName {
			s.profiles.selectedIndex = i
			s.currentUiMode = viewProfileUiMode
			break
		}
	}
}

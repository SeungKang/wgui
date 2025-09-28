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

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// LEFT: Sidebar
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(ctx, gtx)
		}),

		// RIGHT: Main content
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// TOP: scrollable form
				layout.Flexed(1, func(gtx C) D {
					form := []layout.Widget{
						func(gtx C) D { return s.renderSpacer(gtx, unit.Dp(16)) },
						s.formField("Name", s.profileNameEditor, unit.Dp(80)),
						s.formField("Config", s.configEditor, unit.Dp(200)),
					}

					// (optional) handle editor updates while laid out
					s.handleNewProfileEditorUpdates(gtx)

					return material.List(s.theme, s.list).Layout(gtx, len(form), func(gtx C, i int) D {
						return layout.UniformInset(unit.Dp(16)).Layout(gtx, form[i])
					})
				}),

				// BOTTOM: fixed action bar (Save/Cancel + error)
				layout.Rigid(func(gtx C) D {
					// Optional background for the bar
					// paint.Fill(gtx.Ops, color.NRGBA{A:0xFF, R:24, G:24, B:28})

					return layout.Inset{
						Top: unit.Dp(8), Left: unit.Dp(16), Right: unit.Dp(16), Bottom: unit.Dp(16),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							// Buttons row
							layout.Rigid(func(gtx C) D {
								return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return s.renderButton(gtx, "Save", PurpleColor, s.saveButton, func() { s.saveProfile(ctx) })
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
											return s.renderButton(gtx, "Cancel", GreyColor, s.cancelButton, func() {
												if len(s.profiles.profiles) != 0 {
													s.currentUiMode = viewProfileUiMode
												}
											})
										})
									}),
									// spacer to take remaining width (keeps buttons left-aligned)
									layout.Flexed(1, func(gtx C) D { return layout.Spacer{}.Layout(gtx) }),
								)
							}),

							// Error message under buttons
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
									return s.renderErrorMessage(gtx, "this is an error message")
								})
							}),
						)
					})
				}),
			)
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

func (s *State) saveProfile(ctx context.Context) {
	profileName := s.profileNameEditor.Text()
	configContent := s.configEditor.Text()

	// TODO if editProfileUiMode, and name is changed, change the config filename

	if profileName == "" || configContent == "" {
		s.errLogger.Printf("Profile name or config content is empty")
		return
	}

	// Create .wgu directory if it doesn't exist
	if err := os.MkdirAll(s.wguDir, 0755); err != nil {
		s.errLogger.Printf("Error creating .wgu directory: %v", err)
		return
	}

	// Create the config file path
	configPath := filepath.Join(s.wguDir, profileName+".conf")

	// Write the config to file
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		s.errLogger.Printf("Error saving config file: %v", err)
		return
	}

	// Clear the editors
	s.profileNameEditor.SetText("")
	s.configEditor.SetText("")

	// Refresh the profiles list
	err := s.RefreshProfiles(ctx)
	if err != nil {
		s.errLogger.Printf("failed to refresh profile - %v", err)
	}

	// Switch to the newly created profile (it will be in the sorted list, not at the end)
	for i, profile := range s.profiles.profiles {
		if profile.name == profileName {
			s.profiles.selectedIndex = i
			s.currentUiMode = viewProfileUiMode
			break
		}
	}
}
func (s *State) handleNewProfileEditorUpdates(gtx layout.Context) {
	// Profile name editor updates
	for {
		_, ok := s.profileNameEditor.Update(gtx)
		if !ok {
			break
		}
	}

	// Config editor updates
	for {
		_, ok := s.configEditor.Update(gtx)
		if !ok {
			break
		}
	}
}

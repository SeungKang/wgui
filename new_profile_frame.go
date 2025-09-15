package main

import (
	"context"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (s *State) renderNewProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor) // Fill background first

	widgets := []layout.Widget{
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderTextEditor(gtx, s.profileNameEditor, "Enter profile here", unit.Dp(80))
		},
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderTextEditor(gtx, s.configEditor, "Enter config here", unit.Dp(200))
		},
		func(gtx C) D {
			return s.renderButton(ctx, gtx, "Save", func() {
				name := strings.TrimSpace(filepath.Base(s.profileNameEditor.Text()))
				configPath := filepath.Join(s.configDirPath, name+".conf")

				data := s.configEditor.Text()
				if !strings.HasSuffix(data, "\n") {
					data += "\n"
				}

				err := os.WriteFile(configPath, []byte(data), 0600)
				if err != nil {
					// TODO need a message box
					log.Fatal(err)
				}

				s.AddProfile(name)

				s.selectedProfile = len(s.profiles) - 1
			})
		},
		func(gtx C) D {
			return s.renderErrorMessage(gtx, "this is an error message")
		},
	}

	// --- LAYOUT: Sidebar (left) + Main (right) ---
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// LEFT: Sidebar column (fixed width)
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(gtx)
		}),

		// RIGHT: Your existing scrollable content
		layout.Flexed(1, func(gtx C) D {
			return material.List(s.theme, s.list).Layout(gtx, len(widgets), func(gtx C, i int) D {
				return layout.Center.Layout(gtx, func(gtx C) D {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
				})
			})
		}),
	)
}

// Handle editor updates for the new profile frame
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
				log.Printf("error: rendering markdown: %v", err)
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
				log.Printf("error: rendering markdown: %v", err)
			}
		}
	}
}

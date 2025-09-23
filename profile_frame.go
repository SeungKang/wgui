package main

import (
	"context"
	"os"
	"wugui/internal/wguctl"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func (s *State) renderProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor) // Fill background first

	selectedProfile := s.profiles.profiles[s.profiles.selectedProfile]

	wguConfig := wguctl.Config{
		ExePath:    "wgu",
		ConfigPath: selectedProfile.configPath,
	}

	widgets := []layout.Widget{
		func(gtx C) D {
			return s.renderProfileTitle(gtx)
		},
		func(gtx C) D {
			return s.renderPubkey(gtx, selectedProfile.pubkey)
		},
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderLogs(gtx)
		},
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderConnectButton(ctx, gtx, wguConfig)
		},
		func(gtx C) D {
			return s.renderEditButton(ctx, gtx, wguConfig)
		},
		func(gtx C) D {
			return s.renderDeleteButton(ctx, gtx)
		},
		func(gtx C) D {
			return s.renderErrorMessage(gtx, "this is an error message")
		},
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// LEFT: Sidebar column (fixed width)
		layout.Rigid(func(gtx C) D {
			// Vertical stack: Button at top, then sidebar
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Top button
				layout.Rigid(func(gtx C) D {
					in := layout.UniformInset(unit.Dp(8))

					btn := material.Button(s.theme, s.newProfileButton, "New Profile")
					btn.Background = PurpleColor

					for s.newProfileButton.Clicked(gtx) {
						s.frame = "new_profile_frame" // navigate to your new-profile UI
						s.win.Invalidate()            // optional: force redraw
					}

					return in.Layout(gtx, btn.Layout)
				}),

				// Sidebar content
				layout.Flexed(1, func(gtx C) D {
					return s.renderSidebar(ctx, gtx)
				}),
			)
		}),

		// RIGHT: Your existing scrollable content
		layout.Flexed(1, func(gtx C) D {
			return material.List(s.theme, s.list).Layout(gtx, len(widgets), func(gtx C, i int) D {
				return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
			})
		}),
	)
}

func (s *State) renderProfileTitle(gtx layout.Context) layout.Dimensions {
	l := material.H4(s.theme, s.profiles.profiles[s.profiles.selectedProfile].name)
	l.Color = WhiteColor
	l.State = new(widget.Selectable) // makes the text selectable
	return l.Layout(gtx)
}

func (s *State) renderConnectButton(ctx context.Context, gtx layout.Context, config wguctl.Config) layout.Dimensions {
	wguState, lastErr := s.wgu.State()

	var label string
	switch wguState {
	case wguctl.DisconnectingFsmState:
		label = "Disconnecting..."
	case wguctl.DisconnectedFsmState:
		label = "Connect"
	case wguctl.ConnectingFsmState:
		label = "Connecting..."
	case wguctl.ConnectedFsmState:
		label = "Disconnect"
	case wguctl.ErrorFsmState:
		label = "Error"
		_ = lastErr
	}

	return s.renderButton(ctx, gtx, label, PurpleColor, func() {
		switch wguState {
		case wguctl.ConnectedFsmState, wguctl.ConnectingFsmState:
			_ = s.wgu.Disconnect(ctx)
		default:
			_ = s.wgu.Connect(ctx, config)
		}
	})
}

func (s *State) renderEditButton(ctx context.Context, gtx layout.Context, config wguctl.Config) layout.Dimensions {
	return s.renderButton(ctx, gtx, "edit", RedColor, func() {
	})
}

func (s *State) renderDeleteButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return s.renderButton(ctx, gtx, "delete", RedColor, func() {
		configPath := s.profiles.profiles[s.profiles.selectedProfile].configPath
		err := os.Remove(configPath)
		if err != nil {
			s.errLogger.Printf("failed to remove wgu config file: %q - %v", configPath, err)
		}

		s.win.Invalidate()
	})
}

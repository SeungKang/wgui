package main

import (
	"context"
	"image/color"
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
		ExePath:    s.wguExePath,
		ConfigPath: selectedProfile.configPath,
	}

	widgets := []layout.Widget{
		func(gtx C) D {
			return s.renderProfileTitle(gtx)
		},
		func(gtx C) D {
			return s.renderPubkey(gtx, "pubkey: "+selectedProfile.pubkey)
		},
		func(gtx C) D {
			return s.renderLogs(gtx)
		},
		// Connect, Edit, and Delete buttons on the same row
		func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return s.renderConnectButton(ctx, gtx, wguConfig)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
						return s.renderEditButton(ctx, gtx)
					})
				}),
			)
		},
		func(gtx C) D {
			return s.renderErrorMessage(gtx, "this is an error message")
		},
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

func (s *State) renderProfileTitle(gtx layout.Context) layout.Dimensions {
	l := material.H5(s.theme, s.profiles.profiles[s.profiles.selectedProfile].name)
	l.Color = PurpleColor
	l.State = new(widget.Selectable) // makes the text selectable
	return l.Layout(gtx)
}

func (s *State) renderConnectButton(ctx context.Context, gtx layout.Context, config wguctl.Config) layout.Dimensions {
	wguState, lastErr := s.profiles.profiles[s.profiles.selectedProfile].wgu.State()

	var label string
	var color color.NRGBA
	switch wguState {
	case wguctl.DisconnectingFsmState:
		label = "Disconnecting..."
		color = RedColor
	case wguctl.DisconnectedFsmState:
		label = "Connect"
		color = GreenColor
	case wguctl.ConnectingFsmState:
		label = "Connecting..."
		color = GreenColor
	case wguctl.ConnectedFsmState:
		label = "Disconnect"
		color = RedColor
	case wguctl.ErrorFsmState:
		label = "Error"
		color = RedColor
		_ = lastErr
	default:
		// do nothing
	}

	return s.renderButton(gtx, label, color, s.connectButton, func() {
		switch wguState {
		case wguctl.ConnectedFsmState, wguctl.ConnectingFsmState:
			_ = s.profiles.profiles[s.profiles.selectedProfile].wgu.Disconnect(ctx)
		default:
			_ = s.profiles.profiles[s.profiles.selectedProfile].wgu.Connect(ctx, config)
		}
	})
}

func (s *State) renderEditButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return s.renderButton(gtx, "Edit", GreyColor, s.editButton, func() {
		s.currentUiMode = editProfileUiMode

		s.profileNameEditor.SetText(s.profiles.profiles[s.profiles.selectedProfile].name)
		s.configEditor.SetText(s.profiles.profiles[s.profiles.selectedProfile].lastReadConfig)
	})
}

func (s *State) renderDeleteButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return s.renderButton(gtx, "Delete", RedColor, s.deleteButton, func() {
		configPath := s.profiles.profiles[s.profiles.selectedProfile].configPath
		err := os.Remove(configPath)
		if err != nil {
			s.errLogger.Printf("failed to remove wgu config file: %q - %v", configPath, err)
		}

		// Refresh the profiles list
		err = s.RefreshProfiles(ctx)
		if err != nil {
			s.errLogger.Printf("failed to refresh profile - %v", err)
		}
	})
}

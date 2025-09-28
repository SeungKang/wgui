package main

import (
	"context"
	"image/color"
	"os"
	"strings"
	"wugui/internal/wguctl"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func (s *State) renderProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor) // Fill background first

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		// LEFT: Sidebar
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(ctx, gtx)
		}),

		// RIGHT: Main content
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// TOP: fixed header (title + pubkey)
				layout.Rigid(func(gtx C) D {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D { return s.renderProfileTitle(gtx) }),
							layout.Rigid(func(gtx C) D {
								return s.renderPubkey(gtx, "pubkey: "+s.profiles.selected().pubkey)
							}),
						)
					})
				}),

				// MIDDLE: scrollable logs only (List-based)
				layout.Flexed(1, func(gtx C) D {
					logs := s.profiles.selected().wgu.Stderr()
					lines := strings.Split(strings.ReplaceAll(logs, "\r\n", "\n"), "\n")

					// trim trailing empty line
					if len(lines) > 0 && lines[len(lines)-1] == "" {
						lines = lines[:len(lines)-1]
					}

					// 🔹 Grow/shrink logSelectables here
					for len(s.logSelectables) < len(lines) {
						s.logSelectables = append(s.logSelectables, widget.Selectable{})
					}
					if len(s.logSelectables) > len(lines) {
						s.logSelectables = s.logSelectables[:len(lines)]
					}

					return material.List(s.theme, s.logsList).Layout(gtx, len(lines), func(gtx C, i int) D {
						row := material.Body1(s.theme, lines[i])
						row.State = &s.logSelectables[i]
						row.Color = WhiteColor
						row.TextSize = unit.Sp(12)
						row.Font.Typeface = "monospace"

						return layout.Inset{
							Top: unit.Dp(2), Bottom: unit.Dp(0),
							Left: unit.Dp(16), Right: unit.Dp(8),
						}.Layout(gtx, row.Layout)
					})
				}),

				// BOTTOM: fixed bar (buttons + error)
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top: unit.Dp(16), Left: unit.Dp(16), Right: unit.Dp(16), Bottom: unit.Dp(16),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							// Buttons row
							layout.Rigid(func(gtx C) D {
								wguConfig := wguctl.Config{ExePath: s.wguExePath, ConfigPath: s.profiles.selected().configPath}
								return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
									layout.Rigid(func(gtx C) D { return s.renderConnectButton(ctx, gtx, wguConfig) }),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
											return s.renderEditButton(ctx, gtx)
										})
									}),
									layout.Flexed(1, func(gtx C) D { return layout.Spacer{}.Layout(gtx) }),
								)
							}),
							// Error message
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

func (s *State) renderProfileTitle(gtx layout.Context) layout.Dimensions {
	l := material.H5(s.theme, s.profiles.selected().name)
	l.Color = PurpleColor
	l.State = new(widget.Selectable) // makes the text selectable
	return l.Layout(gtx)
}

func (s *State) renderConnectButton(ctx context.Context, gtx layout.Context, config wguctl.Config) layout.Dimensions {
	wguState, lastErr := s.profiles.selected().wgu.State()

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
			_ = s.profiles.selected().wgu.Disconnect(ctx)
		default:
			_ = s.profiles.selected().wgu.Connect(ctx, config)
		}
	})
}

func (s *State) renderEditButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return s.renderButton(gtx, "Edit", GreyColor, s.editButton, func() {
		s.currentUiMode = editProfileUiMode

		s.profileNameEditor.SetText(s.profiles.selected().name)
		s.configEditor.SetText(s.profiles.selected().lastReadConfig)
	})
}

func (s *State) renderDeleteButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return s.renderButton(gtx, "Delete", RedColor, s.deleteButton, func() {
		configPath := s.profiles.selected().configPath
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

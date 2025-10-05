package main

import (
	"context"
	"gioui.org/io/clipboard"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"io"
	"os"
	"strings"
	"time"
	"wugui/internal/wguctl"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// renderProfileFrame is the main layout with sidebar and content area
func (s *State) renderProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor)

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return s.renderSidebar(ctx, gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return s.renderProfileContent(ctx, gtx)
		}),
	)
}

// renderProfileContent contains the header, logs, and action bar
func (s *State) renderProfileContent(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return s.renderProfileHeader(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return s.renderLogsSection(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return s.renderActionBar(ctx, gtx)
		}),
	)
}

// renderProfileHeader shows the profile name and public key with copy button
func (s *State) renderProfileHeader(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return s.renderProfileTitle(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				if s.profiles.selected().lastErrMsg != "" {
					return D{}
				}

				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return s.renderPubkey(gtx, "pubkey: "+s.profiles.selected().pubkey)
					}),
					layout.Rigid(func(gtx C) D {
						return s.renderCopyButton(gtx)
					}),
				)
			}),
		)
	})
}

// renderCopyButton displays an icon button to copy the pubkey
func (s *State) renderCopyButton(gtx layout.Context) layout.Dimensions {
	icon, err := widget.NewIcon(icons.ContentContentCopy)
	if err != nil {
		s.errLogger.Printf("failed to create copy icon: %v", err)
		return layout.Dimensions{}
	}

	if s.copyIconButton.Clicked(gtx) {
		gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(s.profiles.selected().pubkey))})
		s.showCopiedMessage()
	}

	return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				btn := material.IconButton(s.theme, s.copyIconButton, icon, "Copy public key")
				btn.Size = unit.Dp(16)
				btn.Inset = layout.UniformInset(unit.Dp(2))
				btn.Description = "copy public key to clipboard"
				return btn.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				if s.isCopiedMessageVisible() {
					return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
						label := material.Body2(s.theme, "copied!")
						label.Color = LightGreyColor
						label.TextSize = unit.Sp(11)
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),
		)
	})
}

func (s *State) showCopiedMessage() {
	s.copiedMessageTime = time.Now()
	// Schedule a redraw after 2 seconds to hide the message
	go func() {
		time.Sleep(2 * time.Second)
		s.win.Invalidate() // Trigger a redraw
	}()
}

func (s *State) isCopiedMessageVisible() bool {
	return time.Since(s.copiedMessageTime) < 2*time.Second
}

// renderLogsSection displays scrollable logs
func (s *State) renderLogsSection(gtx layout.Context) layout.Dimensions {
	logs := s.profiles.selected().wgu.Stderr()

	return material.List(s.theme, s.logsList).Layout(gtx, 1, func(gtx C, i int) D {
		row := material.Body1(s.theme, logs)
		row.State = s.logSelectables
		row.Color = WhiteColor
		row.TextSize = unit.Sp(12)
		row.Font.Typeface = "monospace"

		return layout.Inset{
			Top: unit.Dp(2), Bottom: unit.Dp(0),
			Left: unit.Dp(16), Right: unit.Dp(8),
		}.Layout(gtx, row.Layout)
	})
}

// renderActionBar contains buttons and error messages
func (s *State) renderActionBar(ctx context.Context, gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{
		Top: unit.Dp(16), Left: unit.Dp(16),
		Right: unit.Dp(16), Bottom: unit.Dp(16),
	}

	return inset.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return s.renderActionButtons(ctx, gtx)
			}),
			layout.Rigid(func(gtx C) D {
				// Reserve consistent space for error message
				minHeight := gtx.Dp(unit.Dp(32)) // adjust as needed (24–32dp looks good)
				gtx.Constraints.Min.Y = minHeight

				return s.renderErrorSection(gtx)
			}),
		)
	})
}

// renderActionButtons shows the Connect and Edit buttons
func (s *State) renderActionButtons(ctx context.Context, gtx layout.Context) layout.Dimensions {
	wguConfig := wguctl.Config{
		ExePath:    s.wguExePath,
		ConfigPath: s.profiles.selected().configPath,
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return s.renderConnectButton(ctx, gtx, wguConfig)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
				return s.renderEditButton(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Spacer{}.Layout(gtx)
		}),
	)
}

// renderErrorSection displays error messages
func (s *State) renderErrorSection(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
		return s.renderErrorMessage(gtx, s.profiles.selected().lastErrMsg)
	})
}

// renderProfileTitle displays the profile name as a heading
func (s *State) renderProfileTitle(gtx layout.Context) layout.Dimensions {
	l := material.H5(s.theme, s.profiles.selected().name)
	l.Color = PurpleColor
	l.State = new(widget.Selectable)
	return l.Layout(gtx)
}

// renderConnectButton shows a button that changes based on connection state
func (s *State) renderConnectButton(ctx context.Context, gtx layout.Context, config wguctl.Config) layout.Dimensions {
	label, color := s.getConnectButtonStyle()
	onClick := s.createConnectButtonHandler(ctx, config)

	return s.renderButton(gtx, label, color, s.connectButton, onClick)
}

// getConnectButtonStyle returns the label and color based on connection state
func (s *State) getConnectButtonStyle() (string, color.NRGBA) {
	selected := s.profiles.selected()
	if selected.lastErrMsg != "" {
		return "Error", RedColor
	}

	wguState, lastErr := selected.wgu.State()
	_ = lastErr

	switch wguState {
	case wguctl.DisconnectingFsmState:
		return "Disconnecting...", RedColor
	case wguctl.DisconnectedFsmState:
		return "Connect", GreenColor
	case wguctl.ConnectingFsmState:
		return "Connecting...", GreenColor
	case wguctl.ConnectedFsmState:
		return "Disconnect", RedColor
	case wguctl.ErrorFsmState:
		return "Error", RedColor
	default:
		return "", color.NRGBA{}
	}
}

// createConnectButtonHandler returns the appropriate click handler
func (s *State) createConnectButtonHandler(ctx context.Context, config wguctl.Config) func() {
	wguState, lastErr := s.profiles.selected().wgu.State()
	_ = lastErr

	return func() {
		switch wguState {
		case wguctl.ConnectedFsmState, wguctl.ConnectingFsmState:
			_ = s.profiles.selected().wgu.Disconnect(ctx)
		default:
			_ = s.profiles.selected().wgu.Connect(ctx, config)
		}
	}
}

// renderEditButton shows the edit profile button
func (s *State) renderEditButton(gtx layout.Context) layout.Dimensions {
	onClick := func() {
		s.switchToEditMode()
	}
	return s.renderButton(gtx, "Edit", GreyColor, s.editButton, onClick)
}

// switchToEditMode changes UI mode and populates editors
func (s *State) switchToEditMode() {
	s.currentUiMode = editProfileUiMode
	s.profileNameEditor.SetText(s.profiles.selected().name)
	s.configEditor.SetText(s.profiles.selected().lastReadConfig)
	s.errLabel = ""
}

// deleteProfile removes the profile config and refreshes the list
func (s *State) deleteProfile(ctx context.Context) {
	configPath := s.profiles.selected().configPath

	if err := os.Remove(configPath); err != nil {
		s.errLogger.Printf("failed to remove wgu config file: %q - %v", configPath, err)
	}

	if err := s.RefreshProfiles(ctx); err != nil {
		s.errLogger.Printf("failed to refresh profile - %v", err)
	}
}

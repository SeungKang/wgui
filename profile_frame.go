package main

import (
	"context"
	"wugui/internal/wguctl"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func (s *State) renderProfileFrame(ctx context.Context, gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, BgColor) // Fill background first

	widgets := []layout.Widget{
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderProfileTitle(gtx)
		},
		func(gtx C) D {
			return s.renderSpacer(gtx, unit.Dp(16))
		},
		func(gtx C) D {
			return s.renderWireguardButton(ctx, gtx)
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

func (s *State) renderProfileTitle(gtx layout.Context) layout.Dimensions {
	l := material.H4(s.theme, s.profiles[s.selectedProfile])
	l.Color = WhiteColor
	l.State = new(widget.Selectable) // makes the text selectable
	return l.Layout(gtx)
}

func (s *State) renderWireguardButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
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

	return s.renderButton(ctx, gtx, label, func() {
		switch wguState {
		case wguctl.ConnectedFsmState, wguctl.ConnectingFsmState:
			_ = s.wgu.Disconnect(ctx)
		default:
			_ = s.wgu.Connect(ctx, wguctl.WguConfig{ConfigPath: "/Users/kang_/.wgu/gamingbsd.conf"})
		}
	})
}

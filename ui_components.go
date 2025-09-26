package main

import (
	"context"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	D = layout.Dimensions
	C = layout.Context
)

func (s *State) renderSidebar(ctx context.Context, gtx layout.Context) layout.Dimensions {
	width := gtx.Dp(unit.Dp(200))
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width

	// Sidebar background
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, SidebarBg, rect)

	in := layout.UniformInset(unit.Dp(8))
	return in.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// New profile button at top
			layout.Rigid(func(gtx C) D {
				btn := material.Button(s.theme, s.newProfileButton, "+")
				btn.Background = PurpleColor

				for s.newProfileButton.Clicked(gtx) {
					s.profileNameEditor.SetText("")
					s.configEditor.SetText("")
					s.currentUiMode = newProfileUiMode
					s.win.Invalidate()
				}

				return btn.Layout(gtx)
			}),

			// Spacing between button and list
			layout.Rigid(func(gtx C) D {
				return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
			}),

			// Profile list below button
			layout.Flexed(1, func(gtx C) D {
				return material.List(s.theme, s.profiles.profileList).Layout(gtx, len(s.profiles.profiles), func(gtx C, i int) D {
					for s.profiles.profileClicks[i].Clicked(gtx) {
						err := s.profiles.profiles[i].refresh(ctx, s.wguExePath)
						if err != nil {
							s.errLogger.Printf("failed to refresh profile %s - %v", s.profiles.profiles[i].name, err)
						}

						s.profiles.selectedProfile = i
						s.currentUiMode = viewProfileUiMode
						s.win.Invalidate()
					}

					// Row styling (highlight selected only when on profile frame)
					row := func(gtx C) D {
						if i == s.profiles.selectedProfile && s.currentUiMode != newProfileUiMode {
							paint.FillShape(gtx.Ops, SelectedBg, clip.Rect{Max: gtx.Constraints.Max}.Op())
						}

						pad := layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(6), Left: unit.Dp(8), Right: unit.Dp(8)}
						return pad.Layout(gtx, func(gtx C) D {
							lbl := material.Body1(s.theme, s.profiles.profiles[i].name)
							lbl.Color = WhiteColor
							return lbl.Layout(gtx)
						})
					}

					return s.profiles.profileClicks[i].Layout(gtx, row)
				})
			}),
		)
	})
}

func (s *State) renderTextEditor(gtx layout.Context, editor *widget.Editor, placeholder string, height unit.Dp) layout.Dimensions {
	h := gtx.Dp(height)
	gtx.Constraints.Min.Y = h
	gtx.Constraints.Max.Y = h

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			// Background color
			rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
			paint.FillShape(gtx.Ops, GreyColor, rect)

			// Apply internal padding
			in := layout.Inset{
				Left:   unit.Dp(8),
				Right:  unit.Dp(8),
				Top:    unit.Dp(6),
				Bottom: unit.Dp(6),
			}

			for {
				_, ok := editor.Update(gtx)
				if !ok {
					break
				}
			}

			return in.Layout(gtx, func(gtx C) D {
				ed := material.Editor(s.theme, editor, placeholder)
				ed.Color = WhiteColor
				return ed.Layout(gtx)
			})
		}),
	)
}

func (s *State) renderButton(ctx context.Context, gtx layout.Context, label string, color color.NRGBA, button *widget.Clickable, onClick func()) layout.Dimensions {
	in := layout.UniformInset(unit.Dp(0))
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return in.Layout(gtx, func(gtx C) D {
				for button.Clicked(gtx) {
					onClick()
				}

				btn := material.Button(s.theme, button, label)
				btn.Background = color
				return btn.Layout(gtx)
			})
		}),
	)
}

func (s *State) renderLogs(gtx layout.Context) layout.Dimensions {
	logsBody := material.Body1(s.theme, "this is a message body")
	logsBody.Color = WhiteColor
	logsBody.Alignment = text.Start
	return logsBody.Layout(gtx)
}

func (s *State) renderErrorMessage(gtx layout.Context, message string) layout.Dimensions {
	errorMessage := material.Label(s.theme, 12, message)
	errorMessage.Color = RedColor
	return errorMessage.Layout(gtx)
}

func (s *State) renderPubkey(gtx layout.Context, message string) layout.Dimensions {
	pubkeyLabel := material.Label(s.theme, 16, message)
	pubkeyLabel.Color = WhiteColor
	return pubkeyLabel.Layout(gtx)
}

func (s *State) renderSpacer(gtx layout.Context, height unit.Dp) layout.Dimensions {
	return layout.Spacer{Height: height}.Layout(gtx)
}

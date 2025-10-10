package main

import (
	"context"
	"image"
	"image/color"

	"golang.org/x/exp/shiny/materialdesign/icons"

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
			layout.Rigid(func(gtx C) D {
				return s.renderSidebarButtons(ctx, gtx)
			}),

			// Spacing between buttons and list
			layout.Rigid(func(gtx C) D {
				return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
			}),

			// Profile list below buttons
			layout.Flexed(1, func(gtx C) D {
				return material.List(s.theme, s.profiles.profileList).Layout(gtx, len(s.profiles.profiles), func(gtx C, i int) D {
					for s.profiles.profileClicks[i].Clicked(gtx) {
						s.profiles.profiles[i].refresh(ctx, s.wguExePath, s.errLogger)

						s.profiles.selectedIndex = i
						s.currentUiMode = viewProfileUiMode
						s.win.Invalidate()
					}

					// Row styling (highlight selected only when on profile frame)
					row := func(gtx C) D {
						if i == s.profiles.selectedIndex && s.currentUiMode != newProfileUiMode {
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

// renderSidebarButtons shows the new profile and refresh buttons
func (s *State) renderSidebarButtons(ctx context.Context, gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		// New profile button (left)
		layout.Rigid(func(gtx C) D {
			btnWidth := gtx.Dp(40)
			gtx.Constraints.Min.X, gtx.Constraints.Max.X = btnWidth, btnWidth

			btn := material.Button(s.theme, s.newProfileButton, "+")
			btn.Background = PurpleColor

			for s.newProfileButton.Clicked(gtx) {
				s.profileNameEditor.SetText("")
				s.configEditor.SetText("")
				s.errLabel = ""
				s.currentUiMode = newProfileUiMode
				s.win.Invalidate()
			}

			return btn.Layout(gtx)
		}),

		// Spacer to push refresh button to the right
		layout.Flexed(1, func(gtx C) D {
			return layout.Spacer{}.Layout(gtx)
		}),

		// Refresh button (right)
		layout.Rigid(func(gtx C) D {
			return s.renderSidebarRefreshButton(ctx, gtx)
		}),
	)
}

// renderSidebarRefreshButton shows the refresh icon button
func (s *State) renderSidebarRefreshButton(ctx context.Context, gtx layout.Context) layout.Dimensions {
	icon, err := widget.NewIcon(icons.NavigationRefresh)
	if err != nil {
		s.errLogger.Printf("failed to create refresh icon: %v", err)
		return layout.Dimensions{}
	}

	if s.refreshIconButton.Clicked(gtx) {
		_ = s.RefreshProfiles(ctx)
	}

	// Match the size of the "+" button
	btnSize := gtx.Dp(40)
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = btnSize, btnSize
	gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = btnSize, btnSize

	return s.refreshIconButton.Layout(gtx, func(gtx C) D {

		// Center the icon
		return layout.Center.Layout(gtx, func(gtx C) D {
			gtx.Constraints.Min = image.Point{}
			return icon.Layout(gtx, GreenColor)
		})
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
				ed.TextSize = unit.Sp(14)
				return ed.Layout(gtx)
			})
		}),
	)
}

func (s *State) renderButton(gtx layout.Context, label string, color color.NRGBA, button *widget.Clickable, onClick func()) layout.Dimensions {
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
	logsBody := material.Label(s.theme, 12, s.profiles.selected().wgu.Stderr())
	logsBody.Color = WhiteColor
	logsBody.Alignment = text.Start

	return logsBody.Layout(gtx)
}

func (s *State) renderErrorMessage(gtx layout.Context, message string) layout.Dimensions {
	errorMessage := material.Label(s.theme, 12, message)
	errorMessage.Color = RedColor
	errorMessage.State = s.errorSelectable
	return errorMessage.Layout(gtx)
}

func (s *State) renderPubkey(gtx layout.Context, message string) layout.Dimensions {
	pubkeyLabel := material.Label(s.theme, 12, message)
	pubkeyLabel.Color = LightGreyColor
	pubkeyLabel.State = s.pubkeySelectable

	return layout.Inset{
		Top: unit.Dp(2),
	}.Layout(gtx, pubkeyLabel.Layout)
}

func (s *State) renderSpacer(gtx layout.Context, height unit.Dp) layout.Dimensions {
	return layout.Spacer{Height: height}.Layout(gtx)
}

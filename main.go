package main

import (
	"context"
	"flag"
	"gioui.org/op/clip"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wugui/internal/wguctl"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	disable = flag.Bool("disable", false, "disable all widgets")
)

type state struct {
	wgu       *wguctl.Fsm
	connected bool
}

var (
	configEditor  = new(widget.Editor)
	connectButton = new(widget.Clickable)
	list          = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
	whiteColor = color.NRGBA{A: 0xff, R: 255, G: 255, B: 255}
	greyColor  = color.NRGBA{A: 0xff, R: 75, G: 75, B: 75}
	redColor   = color.NRGBA{A: 0xff, R: 255, G: 0, B: 0}
	bgColor    = color.NRGBA{A: 0xff, R: 30, G: 30, B: 30}
)

func main() {
	flag.Parse()

	ctx, cancelFn := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancelFn()

	go func() {
		w := new(app.Window)
		w.Option(
			app.Size(unit.Dp(800), unit.Dp(600)),
			app.Title("wugui"),
		)

		s := &state{
			wgu: wguctl.NewFsm(ctx),
		}

		err := loop(ctx, s, w)
		cancelFn()

		select {
		case <-time.After(2 * time.Second):
		case <-s.wgu.Done():
		}

		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func loop(ctx context.Context, s *state, w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	events := make(chan event.Event)
	acks := make(chan struct{})

	go func() {
		for {
			ev := w.Event()
			events <- ev
			<-acks
			if _, ok := ev.(app.DestroyEvent); ok {
				return
			}
		}
	}()

	var ops op.Ops
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-events:
			switch e := e.(type) {
			case app.DestroyEvent:
				acks <- struct{}{}
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				if *disable {
					gtx = gtx.Disabled()
				}

				paint.Fill(gtx.Ops, bgColor) // Fill background first

				wuguiScreen(ctx, s, gtx, th)

				e.Frame(gtx.Ops)
			}
			acks <- struct{}{}
		}
	}
}

type (
	D = layout.Dimensions
	C = layout.Context
)

func wuguiScreen(ctx context.Context, s *state, gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		},
		func(gtx C) D {
			l := material.H4(th, "wugui welcome")
			l.Color = whiteColor
			l.State = new(widget.Selectable) // makes the text selectable
			return l.Layout(gtx)
		},
		func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		},
		func(gtx C) D {
			// Fix the height to 200dp
			h := gtx.Dp(unit.Dp(200))
			gtx.Constraints.Min.Y = h
			gtx.Constraints.Max.Y = h

			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					// Background color
					bg := greyColor
					rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
					paint.FillShape(gtx.Ops, bg, rect)

					// Apply internal padding
					in := layout.Inset{
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
						Top:    unit.Dp(6),
						Bottom: unit.Dp(6),
					}
					return in.Layout(gtx, func(gtx C) D {
						ed := material.Editor(th, configEditor, "Enter config here")
						ed.Color = whiteColor
						return ed.Layout(gtx)
					})
				}),
			)
		},
		func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		},
		func(gtx C) D {
			in := layout.UniformInset(unit.Dp(8))
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return in.Layout(gtx, func(gtx C) D {
						for connectButton.Clicked(gtx) {
							wguState, _ := s.wgu.State()

							switch wguState {
							case wguctl.ConnectedFsmState, wguctl.ConnectingFsmState:
								_ = s.wgu.Disconnect(ctx)
							default:
								_ = s.wgu.Connect(ctx, wguctl.WguConfig{ConfigPath: "/Users/kang_/.wgu/gamingbsd.conf"})
							}
						}

						var label string

						wguState, lastErr := s.wgu.State()

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

						btn := material.Button(th, connectButton, label)
						btn.Background = color.NRGBA{A: 0xff, R: 99, G: 96, B: 225} // purple button
						return btn.Layout(gtx)
					})
				}),
			)
		},
		func(gtx C) D {
			errorMessage := material.Label(th, 12, "this is an error message")
			errorMessage.Color = redColor
			return errorMessage.Layout(gtx)
		},
	}

	return material.List(th, list).Layout(gtx, len(widgets), func(gtx C, i int) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
		})
	})
}

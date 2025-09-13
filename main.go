package main

import (
	"context"
	"flag"
	"image/color"
	"log"
	"os"
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
	screenshot = flag.String("screenshot", "", "save a screenshot to a file and exit")
	disable    = flag.Bool("disable", false, "disable all widgets")
)

type iconAndTextButton struct {
	theme  *material.Theme
	button *widget.Clickable
	icon   *widget.Icon
	word   string
}

type state struct {
	currentScreen string
	gameScreen    *gameScreenState
	wgu           *wguctl.Wgu
	connected     bool
}

type gameScreenState struct {
	cards       []*CardData
	activeCards []*CardData
	cardRows    int
	score       int
	gameWon     bool
}

func (o *gameScreenState) checkWin() bool {
	for _, card := range o.cards {
		if !card.Found {
			return false
		}
	}

	return true
}

func (o *gameScreenState) hideActiveCards() {
	for _, card := range o.cards {
		if !card.Found {
			card.Toggled = false
		}
	}
}

func (o *gameScreenState) resetGame() {
	o.cards = nil
	for i := 0; i < o.cardRows*o.cardRows; i++ {
		o.cards = append(o.cards, &CardData{
			Clickable: new(widget.Clickable),
			Toggled:   false,
		})
	}

	for _, card := range o.cards {
		card.Found = false
		card.Toggled = false
	}

	o.activeCards = nil
	o.score = 0
	o.gameWon = false
}

type CardData struct {
	Clickable *widget.Clickable
	Toggled   bool
	Found     bool
	Shape     string
}

var (
	button                  = new(widget.Clickable)
	radioButtonsPlayerCount = new(widget.Enum)
	radioButtonsGridSize    = new(widget.Enum)
	list                    = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
	whiteColor = color.NRGBA{A: 0xff, R: 255, G: 255, B: 255}
	bgColor    = color.NRGBA{A: 0xff, R: 30, G: 30, B: 30}
)

func main() {
	flag.Parse()

	radioButtonsPlayerCount.Value = "r1" // default
	radioButtonsGridSize.Value = "4"     // default

	go func() {
		w := new(app.Window)
		w.Option(
			app.Size(unit.Dp(800), unit.Dp(600)),
			app.Title("Memory Game"),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	s := state{
		currentScreen: "start",
		gameScreen:    &gameScreenState{},
	}

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

				// Fill background first
				paint.Fill(gtx.Ops, bgColor)

				switch s.currentScreen {
				case "start":
					startScreen(&s, gtx, th)
				}

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

func startScreen(s *state, gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		},
		func(gtx C) D {
			l := material.H4(th, "wugui Welcome") // Memory Game Welcome
			l.Color = whiteColor
			l.State = new(widget.Selectable) // makes the text selectable
			return l.Layout(gtx)
		},
		func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		},
		func(gtx C) D { // button
			in := layout.UniformInset(unit.Dp(8))
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return in.Layout(gtx, func(gtx C) D {
						for button.Clicked(gtx) {
							if !s.connected {
								wgu, err := wguctl.StartWgu(context.Background(), wguctl.WguConfig{ConfigPath: "/home/kang_/.wgu/testing/peer1/wgu.conf"})
								if err != nil {
									return layout.Dimensions{}
								}

								s.wgu = wgu
								s.connected = true
							}

							go func() {

							}()

						}
						btn := material.Button(th, button, "Connect")
						btn.Background = color.NRGBA{A: 0xff, R: 99, G: 96, B: 225} // purple button
						return btn.Layout(gtx)
					})
				}),
			)
		},
	}

	return material.List(th, list).Layout(gtx, len(widgets), func(gtx C, i int) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
		})
	})
}

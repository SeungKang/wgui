package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gioui.org/app"
	"gioui.org/unit"
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

		s := NewState(ctx, w)

		err := s.Run(ctx, w)
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

package main

import (
	"context"
	"flag"
	"gioui.org/app"
	"gioui.org/unit"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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

		timeoutctx, cancelTimeoutFn := context.WithTimeout(ctx, 2*time.Second)
		defer cancelTimeoutFn()

		for _, profile := range s.profiles.profiles {
			select {
			case <-timeoutctx.Done():
			case <-profile.wgu.Done():
			}
		}

		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

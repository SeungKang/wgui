package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gioui.org/app"
	"gioui.org/unit"
)

var version = "dev"

func main() {
	flag.Parse()

	ctx, cancelFn := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancelFn()

	go func() {
		w := new(app.Window)
		w.Option(
			app.Size(unit.Dp(800), unit.Dp(600)),
			app.Title(fmt.Sprintf("wgui [%s]", version)),
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

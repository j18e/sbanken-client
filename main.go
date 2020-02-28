package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
	log "github.com/sirupsen/logrus"
)

func init() {
	if _, err := os.Stat(".env"); err != nil {
		// no .env file present
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env: %v", err)
	}
}

func main() {
	stor := NewStorage()
	cli := NewClient(stor)

	// make sure everything works a first time
	if err := cli.Purchases(); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	srv := &Server{Storage: stor, router: r}
	srv.Routes()

	var g run.Group
	{
		// add the data loader
		ctx, cancel := context.WithCancel(context.Background())
		dur := 6 * time.Hour
		g.Add(func() error {
			ticker := time.NewTicker(dur)
			log.Infof("loading transactions from sbanken every %v", dur)
			defer ticker.Stop()
			return cli.Loop(ctx, ticker)
		}, func(error) {
			cancel()
		})
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return srv.Run(ctx)
		}, func(error) {
			cancel()
		})
	}
	{
		// react to ctrl+c
		ctx, cancel := context.WithCancel(context.Background())
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		g.Add(func() error {
			select {
			case sig := <-sigchan:
				err := run.SignalError{Signal: sig}
				log.Infof("received signal %v", err)
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(error) {
			cancel()
		})
	}

	log.Fatalf("the server was terminated with %v", g.Run())
}

package main

import (
	"context"
	"os"
	"time"

	"github.com/j18e/sbanken-client/pkg/client"
	"github.com/j18e/sbanken-client/pkg/notifications"
	"github.com/j18e/sbanken-client/pkg/server"
	"github.com/j18e/sbanken-client/pkg/storage"
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
	stor := storage.NewStorage()
	cli := client.NewClient(stor)

	// make sure everything works a first time
	if err := cli.Purchases(); err != nil {
		log.Fatal(err)
	}

	notifier := notifications.NewNotifier(stor)

	srv := server.NewServer(stor)
	srv.Routes()

	var g run.Group
	{
		// add the data loader
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return cli.Loop(ctx, 6*time.Hour)
		}, func(error) {
			cancel()
		})
	}
	{
		// add the http server
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return srv.Run(ctx)
		}, func(error) {
			cancel()
		})
	}
	{
		// add the notifier
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return notifier.Run(ctx)
		}, func(error) {
			cancel()
		})
	}

	// react to ctrl+c
	g.Add(run.SignalHandler(context.Background(), os.Interrupt, os.Kill))

	log.Fatalf("the server was terminated with %v", g.Run())
}

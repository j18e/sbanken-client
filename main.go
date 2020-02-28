package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
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
	if debug := os.Getenv("DEBUG"); debug != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	stor := storage.NewStorage()
	// cli := client.NewClient(stor)

	// // make sure everything works a first time
	// if err := cli.Purchases(); err != nil {
	// 	log.Fatal(err)
	// }

	srv := server.NewServer(stor)
	srv.Routes()

	var g run.Group
	// {
	// 	// add the data loader
	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	g.Add(func() error {
	// 		return cli.Loop(ctx, 6*time.Hour)
	// 	}, func(error) {
	// 		cancel()
	// 	})
	// }
	{
		// add the http server
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return srv.Run(ctx)
		}, func(error) {
			cancel()
		})
	}
	// react to ctrl+c
	g.Add(run.SignalHandler(context.Background(), os.Interrupt, os.Kill))

	log.Fatalf("the server was terminated with %v", g.Run())
}

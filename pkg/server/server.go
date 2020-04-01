package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/j18e/sbanken-client/pkg/storage"
	log "github.com/sirupsen/logrus"
)

func NewServer(stor *storage.Storage) *Server {
	r := gin.Default()
	r.Routes()
	return &Server{Storage: stor, router: r}
}

type Server struct {
	Storage *storage.Storage
	router  *gin.Engine
}

func (s *Server) Routes() {
	s.router.LoadHTMLGlob("templates/*")
	s.router.Static("/assets", "./static")
	s.router.StaticFile("/favicon.ico", "./static/favicon.ico")
	s.router.GET("/", s.handlerHome())
	s.router.GET("/spending/:year/:month", s.handlerSpendingMonth())

	// api endpoints
	s.router.GET("api/purchases/:year/:month", s.handlerAPIPurchases())
	s.router.GET("/api/purchase/:purchase", s.handlerAPIPurchase())
	// s.router.PUT("/api/purchase/:purchase", s.handlerPurchase())
	s.router.DELETE("/api/purchase/:purchase", s.handlerAPIPurchaseDelete())
}

func (s *Server) Run(ctx context.Context) error {
	const listenAddr = ":8000"
	errchan := make(chan error, 1)
	go func() {
		log.Infof("http server listening on %s", listenAddr)
		errchan <- s.router.Run(listenAddr)
	}()
	select {
	case err := <-errchan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

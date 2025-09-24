// launching the server, DB, kafka, postgres
package appServer

import (
	"context"
	"crypto/tls"
	"log"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ds124wfegd/WB_L2/18/config"

	"github.com/ds124wfegd/WB_L2/18/internal/service"
	"github.com/ds124wfegd/WB_L2/18/internal/transport"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(cfg *config.Config, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		MaxHeaderBytes:    1 << 20,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.Server.Timeout,
		IdleTimeout:       cfg.Server.Idle_timeout,
		ReadHeaderTimeout: 3 * time.Second,
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
		ErrorLog:          log.New(os.Stderr, "SERVER ERROR: ", log.LstdFlags), // os.Stderr can be replaced with ElsasticSearch in the feature
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func NewServer(cfg *config.Config) {

	logrus.SetFormatter(new(logrus.JSONFormatter))
	services := service.NewCalendar()
	handlers := transport.NewEventHandlers(services)

	srv := new(Server)
	go func() {
		if err := srv.Run(cfg, handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("App Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("App Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

}

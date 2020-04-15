package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Run server main run
func Run() {

	idleConnsClosed := make(chan struct{})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("SERVER_ADDR")),
		Handler:      load(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func(srv *http.Server) {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigint)

		<-sigint

		logrus.Info("received an interrupt signal, shut down the server.")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			logrus.Error("HTTP server Shutdown", err)
		}
		close(idleConnsClosed)
	}(server)

	var g errgroup.Group

	g.Go(func() error {
		logrus.Infof("Starting shorten server on %s", os.Getenv("SERVER_ADDR"))
		return server.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		logrus.Fatal(err)
	}

	<-idleConnsClosed
}

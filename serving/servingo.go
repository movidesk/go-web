package serving

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/movidesk/go-web/logs"
)

// Options Serving options
type Options struct {
	AppAddr    string
	AppHandler http.Handler
}

// Run Serve application
func Run(o Options) error {
	log := logs.Single()

	srv := http.Server{
		Addr:    o.AppAddr,
		Handler: o.AppHandler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(errors.Wrapf(err, "Unable to listen and serve"))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrapf(err, "Unable to shutdown"))
	}
	log.Println("Server exiting")

	return nil
}

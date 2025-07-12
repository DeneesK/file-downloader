package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = time.Second * 1

type Logger interface {
	Infoln(args ...interface{})
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

type APP struct {
	srv *http.Server
	log Logger
}

func NewApp(addr string, log Logger) *APP {
	s := http.Server{
		Addr:    addr,
		Handler: nil,
	}
	return &APP{
		srv: &s,
		log: log,
	}
}

func (a *APP) Run() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)
	defer stop()

	a.log.Infoln("starting application, server listening on", a.srv.Addr)

	go func() {
		err := a.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("failed to start server: %s", err)
		}
	}()

	<-ctx.Done()

	a.log.Infoln("application shutdown process...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error during shutdown: %s", err)
	}
	<-shutdownCtx.Done()
	a.log.Infoln("application and server gracefully stopped")
}

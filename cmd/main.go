package main

import (
	"betapi_server/config/envconfig"
	"context"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ReadTimeout  = 15 * time.Second
	WriteTimeout = 15 * time.Second
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

// stop wait for ^C and close cancel channel
func stop(cancel context.CancelFunc) {
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGINT)
	<-exitCh
	cancel()
}

func main() {
	cfg := envconfig.NewConfig()

	ctx, cancel := context.WithCancel(context.Background())
	go stop(cancel)

	production, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer production.Sync()
	logger := production.Sugar()

	hub := newHub()
	go hub.run()

	liveManager := NewManager(hub, duration(cfg.LiveTimeout), cfg.LiveURL, cfg.LiveJSON)
	defer liveManager.ticker.Stop()

	upcomingManager := NewManager(hub, duration(cfg.UpcomingTimeout), cfg.UpcomingURL, cfg.UpcomingJSON)
	defer upcomingManager.ticker.Stop()

	go liveManager.checkNewMatches(ctx, logger)
	go upcomingManager.checkNewMatches(ctx, logger)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r, logger, hub, liveManager.Body, upcomingManager.Body)
	})

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		logger.Fatal(err)
	}

	server := newServer()
	logger.Infof("server started at address %s", cfg.Addr)

	go func() {
		logger.Warn(server.Serve(listener))
	}()

	<-ctx.Done()
	if err = server.Close(); err != nil && err != http.ErrServerClosed {
		logger.Warnf("incorrect server stop: %v", err)
		return
	}
	logger.Infof("server stopped at address %s", cfg.Addr)
}

// newServer returns new http.Server
func newServer() *http.Server {
	return &http.Server{
		Handler:      nil,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}
}

func duration(timeout int) time.Duration {
	return time.Duration(timeout) * time.Second
}

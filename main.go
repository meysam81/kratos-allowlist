package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/meysam81/x/chimux"
	"github.com/meysam81/x/config"
)

func NewConfig() (*Config, error) {
	defaults := map[string]interface{}{
		"port": "8080",
	}

	c := &Config{}
	_, err := config.NewConfig(config.WithDefaults(defaults), config.WithUnmarshalTo(c))

	if err != nil {
		return nil, err
	}

	return c, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := NewConfig()
	if err != nil {
		log.Fatalf("failed initializing config: %s", err)
	}

	app := NewApp(config)

	router := chimux.NewChi()
	apiv1 := chimux.NewChi(chimux.WithLogger(app.logger), chimux.WithLoggingMiddleware())
	internal := chimux.NewChi(chimux.WithHealthz(), chimux.WithMetrics())
	router.Mount("/v1", apiv1)
	router.Mount("/", internal)

	apiv1.Post("/validate", app.Validate)

	app.logger.Info().Str("port", config.Port).Msg("starting the server...")

	if len(config.AllowedDomains) > 0 {
		app.logger.Info().Strs("domains", config.AllowedDomains).Msg("only allowing the configured domains")
	} else {
		app.logger.Warn().Msg("no allowlist configured. accepting all domains. specify the desired domains with ALLOWED__DOMAINS.")
	}

	s := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	ctxS, cancelS := context.WithCancel(ctx)
	defer cancelS()

	go func() {
		defer cancelS()
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			app.logger.Error().Err(err).Msg("server failed to start")
		}
	}()

	<-ctx.Done()
	stop()

	app.logger.Info().Msg("shutdown signal received. waiting for web to stop...")

	<-ctxS.Done()

	app.logger.Info().Msg("shutdown complete. see you next time.")
}

package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/types"

	"go.uber.org/fx"
)

// startServer determines the deployment mode and starts the appropriate server
func startServer(
	lc fx.Lifecycle,
	cfg *config.Configuration,
	r *gin.Engine,
	log *logger.Logger,
) {

	// Set default deployment mode if not provided
	mode := cfg.Deployment.Mode
	if mode == "" {
		mode = types.ModeLocal
	}

	// Start the appropriate server based on deployment mode
	switch mode {
	case types.ModeLambda:
		startLambdaServer(lc, r, log)
	case types.ModeAPI, types.ModeLocal:
		startAPIServer(lc, r, cfg, log)
	default:
		log.Warn("Unknown deployment mode: %s, defaulting to API server", cfg.Deployment.Mode)
		startAPIServer(lc, r, cfg, log)
	}
}

// startAPIServer starts the HTTP server for local/API deployment modes
func startAPIServer(
	lc fx.Lifecycle,
	r *gin.Engine,
	cfg *config.Configuration,
	log *logger.Logger,
) {
	log.Info("Configuring HTTP server on %s", cfg.Server.Address)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting HTTP server...")

			// Run server in a goroutine to not block fx lifecycle
			go func() {
				if err := r.Run(cfg.Server.Address); err != nil {
					log.Fatalf("Failed to start HTTP server: %v", err)
				}
			}()

			log.Info("HTTP server started successfully on %s", cfg.Server.Address)
			log.Info("Access the API at http://localhost%s", cfg.Server.Address)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server...")
			return nil
		},
	})
}

// CloudOS — main entry point for the kernel binary.
//
// Build:
//
//	go build -o bin/cloudos ./tools/cloudos
//
// Build with version injection:
//
//	go build -ldflags="
//	  -X github.com/cloudos/cloudos/packages/version.Number=v0.1.0
//	  -X github.com/cloudos/cloudos/packages/version.Commit=$(git rev-parse --short HEAD)
//	  -X github.com/cloudos/cloudos/packages/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
//	" -o bin/cloudos ./tools/cloudos
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/api"
	"github.com/cloudos/cloudos/packages/config"
	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/version"
)

func main() {
	configPath := flag.String("config", "cloudos.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		return
	}

	// Load configuration.
	var cfg config.Config
	provider := config.NewYAMLProvider()

	if _, err := provider.Load(*configPath); err != nil {
		// If the config file doesn't exist, use defaults.
		cfg = config.DefaultConfig()
	} else {
		cfg = *provider.Get()
	}

	log := logging.NewLogger(logging.ParseLevel(cfg.Logging.Level))
	log.Info("starting cloudos",
		"version", version.Short(),
		"config", *configPath,
	)

	// Create and boot the kernel.
	k, err := kernel.New(cfg)
	if err != nil {
		log.Error("failed to create kernel", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := k.Boot(ctx); err != nil {
		log.Error("failed to boot kernel", "error", err)
		os.Exit(1)
	}

	log.Info("cloudos booted successfully",
		"state", k.State(),
	)

	// Start the Control Plane API server.
	apiAddr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	apiSrv := api.NewServer(k, apiAddr)

	go func() {
		if err := apiSrv.ListenAndServe(); err != nil {
			log.Error("api server error", "error", err)
		}
	}()

	// Wait for shutdown signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Info("received signal, shutting down", "signal", sig.String())

	// Shut down API server first (stop accepting new requests), then kernel.
	if err := apiSrv.Shutdown(ctx); err != nil {
		log.Error("api server shutdown error", "error", err)
	}

	if err := k.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("cloudos shut down gracefully")
}

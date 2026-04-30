package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/database"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/routes"
)

func main() {
	// ── Configuration ─────────────────────────────────────────────────────────
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Seed demo data in development mode
	if cfg.DevMode {
		if err := database.SeedDemoData(db); err != nil {
			log.Fatalf("seeding demo data failed: %v", err)
		}
	}

	// ── Router ────────────────────────────────────────────────────────────────
	r := gin.New()
	if cfg.DevMode {
		if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
			log.Fatalf("failed to configure trusted proxies: %v", err)
		}
	} else {
		// In production we terminate TLS and forward through nginx; do not trust arbitrary proxy headers.
		if err := r.SetTrustedProxies(nil); err != nil {
			log.Fatalf("failed to configure trusted proxies: %v", err)
		}
	}
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.DevMode))

	// Health check — outside versioned API so load balancers can reach it
	// without auth and without bumping the API version.
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "db": "unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	routes.RegisterRoutes(r, db, cfg)

	// ── HTTP server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.ListenAddr(),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start in a goroutine so we can listen for shutdown signals concurrently.
	go func() {
		log.Printf("model-hub listening on %s (dev=%v)", cfg.ListenAddr(), cfg.DevMode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	// Give in-flight requests up to 15 seconds to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}

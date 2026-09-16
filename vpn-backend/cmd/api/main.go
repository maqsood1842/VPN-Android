package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vpn/backend/internal/auth"
	"github.com/vpn/backend/internal/billing"
	"github.com/vpn/backend/internal/server"
	"github.com/vpn/backend/internal/subscription"
	"github.com/vpn/backend/internal/user"
	"github.com/vpn/backend/internal/wireguard"
	"github.com/vpn/backend/pkg/config"
	"github.com/vpn/backend/pkg/db"
	"github.com/vpn/backend/pkg/jwt"
	"github.com/vpn/backend/pkg/redis"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	redisClient, err := redis.New(&cfg.Redis)
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	} else {
		log.Println("Redis connected")
		defer redisClient.Close()
	}

	userRepo := user.NewRepository(database.Pool)
	serverRepo := server.NewRepository(database.Pool, redisClient)
	subRepo := subscription.NewRepository(database.Pool)
	peerRepo := wireguard.NewRepository(database.Pool)

	authService := auth.NewService(userRepo, &cfg.JWT)
	wgService, err := wireguard.NewService(peerRepo, serverRepo, subRepo, cfg.Database.Host)
	if err != nil {
		log.Printf("Warning: WireGuard service not available (requires root/CAP_NET_ADMIN): %v", err)
	}

	// Billing service (Google Play validation)
	billingConfig := billing.LoadConfigFromEnv()
	billingService := billing.NewService(nil, subRepo) // validator will be nil without credentials
	if billingConfig.CredentialsJSON != "" {
		validator, err := billing.NewGooglePlayValidator(ctx, billingConfig.CredentialsJSON, billingConfig.PackageName, billingConfig.ProductID)
		if err != nil {
			log.Printf("Warning: Google Play validator not available: %v", err)
		} else {
			billingService = billing.NewService(validator, subRepo)
		}
	}

	// Start health checker
	healthChecker, err := server.NewHealthChecker(serverRepo, 30*time.Second)
	if err != nil {
		log.Printf("Warning: Health checker not available: %v", err)
	} else {
		go healthChecker.Start(ctx)
		defer healthChecker.Stop()
	}

	r := gin.Default()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		authHandler := auth.NewHandler(authService)
		authHandler.RegisterRoutes(api)

		protected := api.Group("")
		protected.Use(authMiddleware(&cfg.JWT))
		{
			wgHandler := wireguard.NewHandler(wgService)
			wgHandler.RegisterRoutes(protected)

			serverHandler := server.NewHandler(serverRepo, healthChecker)
			serverHandler.RegisterRoutes(protected)

			billingHandler := billing.NewHandler(billingService, &cfg.JWT)
			billingHandler.RegisterRoutes(protected)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func authMiddleware(cfg *jwt.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		claims, err := jwt.ValidateToken(tokenString, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
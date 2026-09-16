package main

import (
	"context"
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
	"github.com/vpn/backend/pkg/logging"
	"github.com/vpn/backend/pkg/ratelimit"
	"github.com/vpn/backend/pkg/redis"
)

func main() {
	// Initialize privacy-compliant logger
	logging.Init(logging.LevelInfo)
	logger := logging.Get()

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, &cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer database.Close()

	redisClient, err := redis.New(&cfg.Redis)
	if err != nil {
		logger.Warn("Redis not available", map[string]interface{}{"error": err.Error()})
	} else {
		logger.Info("Redis connected", nil)
		defer redisClient.Close()
	}

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRateLimiter(redisClient)

	userRepo := user.NewRepository(database.Pool)
	serverRepo := server.NewRepository(database.Pool, redisClient)
	subRepo := subscription.NewRepository(database.Pool)
	peerRepo := wireguard.NewRepository(database.Pool)

	authService := auth.NewService(userRepo, &cfg.JWT)
	wgService, err := wireguard.NewService(peerRepo, serverRepo, subRepo, cfg.Database.Host)
	if err != nil {
		logger.Warn("WireGuard service not available", map[string]interface{}{"error": err.Error()})
	}

	// Billing service (Google Play validation)
	billingConfig := billing.LoadConfigFromEnv()
	billingService := billing.NewService(nil, subRepo)
	if billingConfig.CredentialsJSON != "" {
		validator, err := billing.NewGooglePlayValidator(ctx, billingConfig.CredentialsJSON, billingConfig.PackageName, billingConfig.ProductID)
		if err != nil {
			logger.Warn("Google Play validator not available", map[string]interface{}{"error": err.Error()})
		} else {
			billingService = billing.NewService(validator, subRepo)
		}
	}

	// Start health checker
	healthChecker, err := server.NewHealthChecker(serverRepo, 30*time.Second)
	if err != nil {
		logger.Warn("Health checker not available", map[string]interface{}{"error": err.Error()})
	} else {
		go healthChecker.Start(ctx)
		defer healthChecker.Stop()
	}

	r := gin.Default()
	r.Use(gin.Recovery())

	// Add request logging middleware (privacy-compliant)
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		logger.Info("http_request", map[string]interface{}{
			"method":      method,
			"path":        path,
			"status":      c.Writer.Status(),
			"latency_ms":  time.Since(start).Milliseconds(),
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		authHandler := auth.NewHandler(authService)

		// Apply rate limiting to auth endpoints
		auth := api.Group("/auth")
		auth.Use(rateLimiter.Middleware("auth:signup"))
		auth.POST("/signup", authHandler.Signup)

		authLogin := api.Group("/auth")
		authLogin.Use(rateLimiter.Middleware("auth:login"))
		authLogin.POST("/login", authHandler.Login)

		authRefresh := api.Group("/auth")
		authRefresh.Use(rateLimiter.Middleware("auth:refresh"))
		authRefresh.POST("/refresh", authHandler.Refresh)

		protected := api.Group("")
		protected.Use(authMiddleware(&cfg.JWT))
		{
			// Rate limit wireguard config endpoint
			wg := protected.Group("/wireguard")
			wg.Use(rateLimiter.Middleware("wireguard:config"))
			wgHandler := wireguard.NewHandler(wgService)
			wgHandler.RegisterRoutes(wg)

			// Rate limit servers endpoint
			servers := protected.Group("/servers")
			servers.Use(rateLimiter.Middleware("servers:list"))
			serverHandler := server.NewHandler(serverRepo, healthChecker)
			serverHandler.RegisterRoutes(servers)

			// Rate limit billing endpoints
			billingGroup := protected.Group("/billing")
			billingGroup.Use(rateLimiter.Middleware("billing:validate"))
			billingHandler := billing.NewHandler(billingService, &cfg.JWT)
			billingHandler.RegisterRoutes(billingGroup)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("Starting server", map[string]interface{}{"port": cfg.Server.Port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server", nil)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	logger.Info("Server exited", nil)
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
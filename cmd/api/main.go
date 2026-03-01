package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"subscription_service/internal/config"
	"subscription_service/internal/handler"
	"subscription_service/internal/repository"
	"subscription_service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	setupLogging(cfg.LogLevel)

	if err := runMigrations(cfg); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

	db, err := repository.NewPostgresConnection(cfg.GetPostgresDSN())
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	subRepo := repository.NewSubscriptionRepository(db)
	subService := service.NewSubscriptionService(subRepo)
	subHandler := handler.NewSubscriptionHandler(subService)

	router := setupRouter(subHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		logrus.Infof("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logrus.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
}

func runMigrations(cfg *config.Config) error {
	m, err := migrate.New(
		cfg.MigrationsPath,
		cfg.GetPostgresDSN(),
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logrus.Info("Migrations completed successfully")
	return nil
}

func setupRouter(subHandler *handler.SubscriptionHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		subscriptions := v1.Group("/subscriptions")
		{
			subscriptions.POST("/", subHandler.CreateSubscription)
			subscriptions.GET("/", subHandler.ListSubscriptions)
			subscriptions.GET("/aggregate", subHandler.AggregateSubscriptions)
			subscriptions.GET("/:id", subHandler.GetSubscription)
			subscriptions.PUT("/:id", subHandler.UpdateSubscription)
			subscriptions.DELETE("/:id", subHandler.DeleteSubscription)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

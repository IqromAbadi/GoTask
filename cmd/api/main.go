package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/iqromabadi/gotask/internal/activity"
	"github.com/iqromabadi/gotask/internal/auth"
	"github.com/iqromabadi/gotask/internal/comment"
	"github.com/iqromabadi/gotask/internal/dashboard"
	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/config"
	"github.com/iqromabadi/gotask/internal/platform/database"
	"github.com/iqromabadi/gotask/internal/platform/logger"
	"github.com/iqromabadi/gotask/internal/platform/response"
	"github.com/iqromabadi/gotask/internal/progress"
	"github.com/iqromabadi/gotask/internal/review"
	"github.com/iqromabadi/gotask/internal/task"
	"github.com/iqromabadi/gotask/internal/tasklist"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.AppEnv)
	slog.SetDefault(log)

	log.Info("starting GoTask API", slog.String("env", cfg.AppEnv))

	// Connect to PostgreSQL
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to PostgreSQL")

	// Build router
	router := buildRouter(cfg, db, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("server started", slog.Int("port", cfg.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	log.Info("server stopped")
}

func buildRouter(cfg *config.Config, db *sql.DB, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.CleanPath)
	r.Use(chimw.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.CorsAllowedOrigins))
	r.Use(middleware.BodyLimit)
	r.Use(middleware.ContentType)

	// Rate limiter for auth endpoints
	authLimiter := middleware.NewRateLimiter(5, 10)

	// Health check
	r.Get("/health", healthHandler(db))

	// Initialize auth module
	authRepo := auth.NewPostgresRepository(db)
	tokenService := auth.NewTokenService(
		cfg.JwtAccessSecret,
		cfg.JwtRefreshSecret,
		cfg.JwtAccessTTL,
		cfg.JwtRefreshTTL,
	)
	authService := auth.NewService(authRepo, tokenService, cfg.JwtRefreshTTL)
	authHandler := auth.NewHandler(authService, log)

	// Initialize activity module
	activityRepo := activity.NewPostgresRepository(db)
	activityService := activity.NewService(activityRepo)
	activityHandler := activity.NewHandler(activityService, log)

	// Initialize task list module
	taskListRepo := tasklist.NewPostgresRepository(db)
	taskListService := tasklist.NewService(taskListRepo)

	// Initialize progress module
	progressRepo := progress.NewPostgresRepository(db)
	progressService := progress.NewService(progressRepo, db, activityService)
	progressHandler := progress.NewHandler(progressService, log)

	// Initialize task module (with activity logging)
	taskRepo := task.NewPostgresRepository(db)
	taskService := task.NewService(taskRepo, activityService, db)
	taskHandler := task.NewHandler(taskService, log)

	// Initialize task list handler (depends on task service for board)
	taskListHandler := tasklist.NewHandler(taskListService, taskService, log)

	// Initialize review module
	reviewRepo := review.NewPostgresRepository(db)
	reviewService := review.NewService(reviewRepo, db, activityService)
	reviewHandler := review.NewHandler(reviewService, log)

	// Initialize comment module
	commentRepo := comment.NewPostgresRepository(db)
	commentService := comment.NewService(commentRepo, activityService)
	commentHandler := comment.NewHandler(commentService, log)

	// Initialize dashboard module
	dashboardService := dashboard.NewService(db)
	dashboardHandler := dashboard.NewHandler(dashboardService, log)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes (rate limited)
		r.Group(func(r chi.Router) {
			r.Use(authLimiter.Limit)
			authHandler.RegisterPublicRoutes(r)
		})

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JwtAccessSecret))
			authHandler.RegisterProtectedRoutes(r)
			taskListHandler.RegisterRoutes(r)
			taskHandler.RegisterRoutes(r)
			progressHandler.RegisterRoutes(r)
			activityHandler.RegisterRoutes(r)
			reviewHandler.RegisterRoutes(r)
			commentHandler.RegisterRoutes(r)
			dashboardHandler.RegisterRoutes(r)
		})
	})

	return r
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, response.APIResponse{
				Success: false,
				Message: "Database tidak tersedia",
			})
			return
		}

		response.JSON(w, http.StatusOK, response.APIResponse{
			Success: true,
			Message: "OK",
			Data:    map[string]string{"status": "ok"},
		})
	}
}

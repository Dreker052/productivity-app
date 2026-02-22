package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dreker052/productivity-app/internal/config"
	"github.com/Dreker052/productivity-app/internal/database"
	authHandler "github.com/Dreker052/productivity-app/internal/handler/auth"
	dailyTaskHandler "github.com/Dreker052/productivity-app/internal/handler/dailyTask"
	diaryEntryHandler "github.com/Dreker052/productivity-app/internal/handler/diaryEntry"
	telegramHandler "github.com/Dreker052/productivity-app/internal/handler/telegram"
	userHandler "github.com/Dreker052/productivity-app/internal/handler/user"
	yearlyGoalHandler "github.com/Dreker052/productivity-app/internal/handler/yearlyGoal"
	"github.com/Dreker052/productivity-app/internal/logger"
	"github.com/Dreker052/productivity-app/internal/middleware"
	dailyTaskRepository "github.com/Dreker052/productivity-app/internal/repository/dailyTask"
	diaryEntryRepository "github.com/Dreker052/productivity-app/internal/repository/diaryEntry"
	telegramRepository "github.com/Dreker052/productivity-app/internal/repository/telegram"
	userRepository "github.com/Dreker052/productivity-app/internal/repository/user"
	yearlyGoalRepository "github.com/Dreker052/productivity-app/internal/repository/yearlyGoal"
	authService "github.com/Dreker052/productivity-app/internal/service/auth"
	dailyTaskService "github.com/Dreker052/productivity-app/internal/service/dailyTask"
	diaryEntryService "github.com/Dreker052/productivity-app/internal/service/diaryEntry"
	telegramService "github.com/Dreker052/productivity-app/internal/service/telegram"
	userService "github.com/Dreker052/productivity-app/internal/service/user"
	yearlyGoalService "github.com/Dreker052/productivity-app/internal/service/yearlyGoal"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger := logger.SetupLogger(cfg)
	slog.SetDefault(logger)

	logger.Info("Starting app")

	db := database.InitDB(cfg, logger)

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisAddr,
	}

	queueClient := asynq.NewClient(redisOpt)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.PrometheusMiddleware())

	dailyTaskRepo := dailyTaskRepository.NewDailyTaskRepository(db, logger)
	diaryEntryRepo := diaryEntryRepository.NewDiaryEntryRepository(db, logger)
	userRepo := userRepository.NewUserRepository(db, logger)
	yearlyGoalRepo := yearlyGoalRepository.NewYearlyGoalRepository(db, logger)
	telegramRepo := telegramRepository.NewTelegramRepository(db, logger)

	dailyTaskServ := dailyTaskService.NewDailyTaskService(dailyTaskRepo, logger)
	diaryEntryServ := diaryEntryService.NewDiaryEntryService(diaryEntryRepo, logger)
	authServ := authService.NewAuthService(userRepo, logger, cfg.JWTSecret, cfg.ServerAddr, queueClient)
	userServ := userService.NewUserService(userRepo, logger)
	yearlyGoalServ := yearlyGoalService.NewYearlyGoalService(yearlyGoalRepo, logger)

	telegramServ, err := telegramService.NewTelegramService(cfg.TelegramBotToken, telegramRepo, dailyTaskRepo, logger)
	if err != nil {
		logger.Error("Failed to initialize Telegram service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	dailyTaskHand := dailyTaskHandler.NewDailyTaskHandler(dailyTaskServ, logger)
	diaryEntryHand := diaryEntryHandler.NewDiaryEntryHandler(diaryEntryServ, logger)
	authHand := authHandler.NewAuthHandler(authServ, logger)
	userHand := userHandler.NewUserHandler(userServ, logger)
	yearlyGoalHand := yearlyGoalHandler.NewYearlyGoalHandler(yearlyGoalServ, logger)
	telegramHand := telegramHandler.NewTelegramHandler(telegramServ, dailyTaskServ, logger)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHand.Register)
			authGroup.POST("/login", authHand.Login)
			authGroup.POST("/refresh", authHand.RefreshTokens)
			authGroup.GET("/verify-email", authHand.VerifyEmail)
		}
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/tasks", dailyTaskHand.Create)
		protected.GET("/tasks", dailyTaskHand.GetTasks)
		protected.PATCH("/tasks/:id/toggle", dailyTaskHand.ToggleTask)
		protected.DELETE("/tasks/:id", dailyTaskHand.Delete)
		protected.PATCH("tasks/:id", dailyTaskHand.Update)

		protected.POST("/diary", diaryEntryHand.Save)
		protected.GET("/diary", diaryEntryHand.GetByDate)

		protected.GET("/groups", yearlyGoalHand.GetGroups)
		protected.POST("/groups", yearlyGoalHand.CreateGroup)
		protected.POST("/goals", yearlyGoalHand.CreateGoal)
		protected.PATCH("/goals/:id/progress", yearlyGoalHand.UpdateProgress)

		protected.GET("/telegram/link", telegramHand.GetTelegramLink)
		protected.POST("/share/telegram", telegramHand.ShareDailyTasksToTelegram)

		protected.GET("user/me", userHand.GetUserData)
		protected.PATCH("user/change-name", userHand.ChangeName)
		protected.PATCH("user/change-email", userHand.ChangeEmail)
		protected.PUT("user/change-password", userHand.ChangePassword)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	go func() {
		if err := telegramServ.Start(ctx); err != nil {
			logger.Error("Telegram bot error", slog.String("error", err.Error()))
		}
	}()

	go func() {
		logger.Info("Server listening", slog.String("addr", cfg.ServerAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	logger.Info("Stopping telegram bot...")
	telegramServ.Stop()

	logger.Info("Closing database connection...")
	db.Close()
	if err := queueClient.Close(); err != nil {
		logger.Error("Failed to close queue client", slog.String("error", err.Error()))
	}

	logger.Info("Server exited properly")

}

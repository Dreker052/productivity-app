package main

import (
	"log/slog"
	"os"

	"github.com/Dreker052/productivity-app.git/internal/config"
	"github.com/Dreker052/productivity-app.git/internal/database"
	authHandler "github.com/Dreker052/productivity-app.git/internal/handler/auth"
	dailyTaskHandler "github.com/Dreker052/productivity-app.git/internal/handler/dailyTask"
	diaryEntryHandler "github.com/Dreker052/productivity-app.git/internal/handler/diaryEntry"
	userHandler "github.com/Dreker052/productivity-app.git/internal/handler/user"
	yearlyGoalHandler "github.com/Dreker052/productivity-app.git/internal/handler/yearlyGoal"
	"github.com/Dreker052/productivity-app.git/internal/logger"
	"github.com/Dreker052/productivity-app.git/internal/middleware"
	dailyTaskRepository "github.com/Dreker052/productivity-app.git/internal/repository/dailyTask"
	diaryEntryRepository "github.com/Dreker052/productivity-app.git/internal/repository/diaryEntry"
	userRepository "github.com/Dreker052/productivity-app.git/internal/repository/user"
	yearlyGoalRepository "github.com/Dreker052/productivity-app.git/internal/repository/yearlyGoal"
	authService "github.com/Dreker052/productivity-app.git/internal/service/auth"
	dailyTaskService "github.com/Dreker052/productivity-app.git/internal/service/dailyTask"
	diaryEntryService "github.com/Dreker052/productivity-app.git/internal/service/diaryEntry"
	userService "github.com/Dreker052/productivity-app.git/internal/service/user"
	yearlyGoalService "github.com/Dreker052/productivity-app.git/internal/service/yearlyGoal"
	"github.com/gin-gonic/gin"
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
	defer db.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))

	dailyTaskRepo := dailyTaskRepository.NewDailyTaskRepository(db, logger)
	diaryEntryRepo := diaryEntryRepository.NewDiaryEntryRepository(db, logger)
	userRepo := userRepository.NewUserRepository(db, logger)
	yearlyGoalRepo := yearlyGoalRepository.NewYearlyGoalRepository(db, logger)

	dailyTaskServ := dailyTaskService.NewDailyTaskService(dailyTaskRepo, logger)
	diaryEntryServ := diaryEntryService.NewDiaryEntryService(diaryEntryRepo, logger)
	authServ := authService.NewAuthService(userRepo, logger, cfg.JWTSecret)
	userServ := userService.NewUserService(userRepo, logger)
	yearlyGoalServ := yearlyGoalService.NewYearlyGoalService(yearlyGoalRepo, logger)

	dailyTaskHand := dailyTaskHandler.NewDailyTaskHandler(dailyTaskServ, logger)
	diaryEntryHand := diaryEntryHandler.NewDiaryEntryHandler(diaryEntryServ, logger)
	authHand := authHandler.NewAuthHandler(authServ, logger)
	userHand := userHandler.NewUserHandler(userServ, logger)
	yearlyGoalHand := yearlyGoalHandler.NewYearlyGoalHandler(yearlyGoalServ, logger)

	_ = userHand

	api := r.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHand.Register)
			authGroup.POST("/login", authHand.Login)
			authGroup.POST("/refresh", authHand.RefreshTokens)
		}
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/tasks", dailyTaskHand.Create)
		protected.GET("/tasks", dailyTaskHand.GetTasks)
		protected.PATCH("/tasks/:id/toggle", dailyTaskHand.ToggleTask)

		protected.POST("/diary", diaryEntryHand.Save)
		protected.GET("/diary", diaryEntryHand.GetByDate)

		protected.GET("/groups", yearlyGoalHand.GetGroups)
		protected.POST("/groups", yearlyGoalHand.CreateGroup)
		protected.POST("/goals", yearlyGoalHand.CreateGoal)
		protected.PATCH("/goals/:id/progress", yearlyGoalHand.UpdateProgress)
	}

	if err = r.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}

}

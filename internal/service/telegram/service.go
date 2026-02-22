package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Dreker052/productivity-app/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type telegramService struct {
	bot          *tgbotapi.BotAPI
	telegramRepo repository.TelegramRepository
	taskRepo     repository.DailyTaskRepository
	logger       *slog.Logger
}

func NewTelegramService(token string, telegramRepo repository.TelegramRepository, taskRepo repository.DailyTaskRepository, logger *slog.Logger) (*telegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	return &telegramService{
		bot:          bot,
		telegramRepo: telegramRepo,
		taskRepo:     taskRepo,
		logger:       logger,
	}, nil
}

// 1. Генерация ссылки
func (s *telegramService) GenerateLink(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := strings.ToUpper(uuid.New().String())

	err := s.telegramRepo.SaveLinkToken(ctx, token, userID, 5*time.Minute)
	if err != nil {
		s.logger.Error("Failed to save link token", slog.String("error", err.Error()))
		return "", err
	}

	// Формат: https://t.me/BotName?start=TOKEN
	return fmt.Sprintf("https://t.me/%s?start=%s", s.bot.Self.UserName, token), nil
}

func (s *telegramService) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	s.logger.Info("Telegram bot started")

	for update := range updates {
		if update.Message != nil {

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				args := update.Message.CommandArguments()
				if args != "" {
					s.handleLinking(update.Message.Chat.ID, update.Message.From.UserName, args)
				}
			}

			if update.Message.IsCommand() && update.Message.Command() == "connect" {
				args := update.Message.CommandArguments()
				if args != "" {
					s.handleLinking(update.Message.Chat.ID, update.Message.From.UserName, args)
				}
			}
		}

		if update.ChannelPost != nil {
			if update.ChannelPost.IsCommand() && update.ChannelPost.Command() == "connect" {
				token := update.ChannelPost.CommandArguments()

				if token != "" {
					s.handleLinking(update.ChannelPost.Chat.ID, update.ChannelPost.Chat.Title, token)
				}
			}
		}
	}
	return nil
}

func (s *telegramService) Stop() {
	s.bot.StopReceivingUpdates()
}

// Привязка аккаунта
func (s *telegramService) handleLinking(chatID int64, username string, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := s.telegramRepo.GetUserIDByToken(ctx, token)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка или срок действия ссылки истек.")
		if _, err := s.bot.Send(msg); err != nil {
			s.logger.Error("Failed to send error message", slog.String("error", err.Error()))
		}
		return
	}

	err = s.telegramRepo.SaveIntegration(ctx, userID, chatID, username)
	if err != nil {
		s.logger.Error("Failed to save integration", slog.String("error", err.Error()))
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Аккаунт успешно привязан! Теперь вы можете делиться планами из приложения.")
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("Failed to send message", slog.String("error", err.Error()))
	}
}

func (s *telegramService) SendDailyPlan(userID string, planText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatID, err := s.telegramRepo.GetChatID(ctx, userID)
	if err != nil {
		return fmt.Errorf("telegram not linked")
	}

	msg := tgbotapi.NewMessage(chatID, planText)

	msg.ParseMode = "HTML"

	msg.DisableWebPagePreview = true

	_, err = s.bot.Send(msg)
	return err
}

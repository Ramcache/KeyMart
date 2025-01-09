package handlers

import (
	"KeyMart/internal/services"
	"KeyMart/internal/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func StartBot(db *pgxpool.Pool, logger *utils.Logger) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("Не указан токен бота.")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Panic(err)
	}

	bot.Debug = true
	logger.Info("Бот запущен: " + bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Panic(err)
	}

	for update := range updates {
		if update.Message != nil {
			if update.Message.SuccessfulPayment != nil {
				services.HandleSuccessfulPayment(update, bot, db, logger)
			} else {
				HandleMessage(update, bot, db, logger)
			}
		} else if update.CallbackQuery != nil {
			HandleCallbackQuery(update.CallbackQuery, bot, db, logger)
		} else if update.PreCheckoutQuery != nil {
			services.HandlePreCheckoutQuery(update.PreCheckoutQuery, bot, logger)
		} else {
			logger.Warn(fmt.Sprintf("Неизвестный тип обновления: %+v", update))
		}
	}
}

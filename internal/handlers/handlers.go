package handlers

import (
	"KeyMart/internal/services"
	"KeyMart/internal/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func HandleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	if update.Message != nil && update.Message.Document != nil {
		logger.Info("Обнаружен файл в сообщении")
		services.AddBulkKeysFromFile(update, bot, db, logger)
		return
	}

	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		services.HandleSuccessfulPayment(update, bot, db, logger)
		return
	}

	if update.Message != nil && update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			showMainMenu(update, bot)
		case "helpadmin":
			handleHelpAdmin(update, bot, logger)
		case "help":
			handleHelp(update, bot, logger)
		case "addkey":
			services.AddKey(update, bot, db, logger)
		case "allkeys":
			services.AllKeys(update, bot, db, logger)
		case "addbulk":
			logger.Info("Получены ключи через текстовое сообщение")
			services.AddBulkKeys(update, bot, db, logger)
		case "buykey":
			services.BuyKeyFromMenu(&tgbotapi.CallbackQuery{
				From: update.Message.From,
				Message: &tgbotapi.Message{
					Chat: update.Message.Chat,
				},
			}, bot, db, logger)
		case "mykeys":
			services.MyKeysFromMenu(&tgbotapi.CallbackQuery{
				From: update.Message.From,
				Message: &tgbotapi.Message{
					Chat: update.Message.Chat,
				},
			}, bot, db, logger)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Попробуйте /help для списка доступных команд.")
			bot.Send(msg)
		}
		return
	}

	if update.Message != nil {
		logger.Info(fmt.Sprintf("Получено сообщение без команды: %s", update.Message.Text))
	}
}

func handleHelpAdmin(update tgbotapi.Update, bot *tgbotapi.BotAPI, logger *utils.Logger) {
	adminID := os.Getenv("ADMIN_TELEGRAM_ID")
	if fmt.Sprintf("%d", update.Message.From.ID) != adminID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для выполнения этой команды.")
		bot.Send(msg)
		logger.Warn(fmt.Sprintf("Пользователь %d попытался использовать команду /help", update.Message.From.ID))
		return
	}

	helpText := `
Доступные команды:

/start - Начало работы с ботом.
/helpadmin - Показать это сообщение с описанием команд.
/addkey - Добавить новый ключ (только для администратора).
/addbulk - Добавить несколько ключей (только для администратора). Можно отправить текстовый файл с ключами.
/buykey - Купить ключ. Ключ будет действителен 1 год.
/mykeys - Показать список ваших купленных ключей.
/allkeys - Показать все ключи (проданные и доступные). (только для администратора)
`

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	bot.Send(msg)
}

func handleHelp(update tgbotapi.Update, bot *tgbotapi.BotAPI, logger *utils.Logger) {
	helpText := `
Доступные команды:

/start - Начало работы с ботом.
/help - Показать это сообщение с описанием команд.
/buykey - Купить VPN-ключ. Ключ действителен 1 год.
/mykeys - Показать список ваших купленных ключей.
`

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

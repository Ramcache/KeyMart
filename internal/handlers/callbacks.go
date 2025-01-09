package handlers

import (
	"KeyMart/internal/services"
	"KeyMart/internal/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleCallbackQuery(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := bot.AnswerCallbackQuery(callback); err != nil {
		logger.Error("Ошибка обработки CallbackQuery: " + err.Error())
	}

	switch query.Data {
	case "buy_key":
		services.BuyKeyFromMenu(query, bot, db, logger)
	case "my_keys":
		services.MyKeysFromMenu(query, bot, db, logger)
	case "help":
		showHelpMessage(query, bot)
	case "main_menu":
		showMainMenuCallback(query, bot)

	default:
		logger.Warn("Неизвестная CallbackQuery: " + query.Data)
	}
}

func showHelpMessage(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	helpText := `
Доступные команды:

- Купить VPN-ключ: Нажмите кнопку "Купить VPN-ключ".
- Просмотреть мои ключи: Нажмите кнопку "Мои ключи".
- Помощь: Нажмите кнопку "Помощь".
`

	// Кнопки для возврата в главное меню
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться в меню", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, helpText)
	msg.ReplyMarkup = buttons
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func showMainMenuCallback(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Купить VPN-ключ", "buy_key"),
			tgbotapi.NewInlineKeyboardButtonData("Мои ключи", "my_keys"),
		),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

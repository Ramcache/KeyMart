package handlers

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func showMainMenu(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Купить VPN-ключ", "buy_key"),
			tgbotapi.NewInlineKeyboardButtonData("Мои ключи", "my_keys"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Выберите действие:")
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

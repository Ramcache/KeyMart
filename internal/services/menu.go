package services

import (
	"KeyMart/internal/utils"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
	"time"
)

func BuyKeyFromMenu(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	var keyID int
	var keyValue string
	err := db.QueryRow(context.Background(),
		"SELECT id, key_value FROM vpn_keys WHERE is_sold = FALSE LIMIT 1").Scan(&keyID, &keyValue)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "К сожалению, сейчас нет доступных ключей для покупки.")
		bot.Send(msg)
		return
	}

	providerToken := os.Getenv("YU_KASSA_PROVIDER_TOKEN")
	title := "Ключ FreemanVPN"
	description := "Доступ к VPN на 1 год."
	payload := fmt.Sprintf("user_%d_key_%d", userID, keyID)
	currency := "RUB"
	price := tgbotapi.LabeledPrice{Label: "К оплате", Amount: 60000}
	prices := []tgbotapi.LabeledPrice{price}

	invoice := tgbotapi.NewInvoice(chatID, title, description, payload, providerToken, "vpn_subscription", currency, &prices)
	_, err = bot.Send(invoice)
	if err != nil {
		logger.Error("Ошибка отправки счета: " + err.Error())
		msg := tgbotapi.NewMessage(chatID, "Ошибка при создании счета. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	_, err = db.Exec(context.Background(),
		"INSERT INTO purchases (user_id, key_id, status) VALUES ((SELECT id FROM users WHERE telegram_id = $1), $2, 'pending')",
		userID, keyID)
	if err != nil {
		logger.Error("Ошибка регистрации покупки: " + err.Error())
	}
}

func MyKeysFromMenu(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	var dbUserID int
	err := db.QueryRow(context.Background(), "SELECT id FROM users WHERE telegram_id=$1", userID).Scan(&dbUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			msg := tgbotapi.NewMessage(chatID, "Вы еще не зарегистрированы в системе. Купите ключ, чтобы начать.")
			bot.Send(msg)
			return
		}
		logger.Error(fmt.Sprintf("Ошибка подключения к базе данных: %v", err))
		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при обработке вашего запроса. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	rows, err := db.Query(context.Background(),
		`SELECT vk.key_value, p.expires_at
		 FROM purchases p
		 INNER JOIN vpn_keys vk ON p.key_id = vk.id
		 WHERE p.user_id = $1 AND p.status = 'paid'`, dbUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка получения ключей пользователя: %v", err))
		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при получении ваших ключей. Попробуйте позже.")
		bot.Send(msg)
		return
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		var expiresAt time.Time
		err := rows.Scan(&key, &expiresAt)
		if err != nil {
			logger.Error(fmt.Sprintf("Ошибка сканирования строки: %v", err))
			continue
		}
		keys = append(keys, fmt.Sprintf("*Ключ:* `%s`\n*Действителен до:* `%s`", key, expiresAt.Format("02.01.2006")))
	}

	if len(keys) == 0 {
		msg := tgbotapi.NewMessage(chatID, "У вас пока нет купленных ключей.")
		bot.Send(msg)
		return
	}

	message := "*Ваши ключи:*\n\n" + strings.Join(keys, "\n\n")
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

package services

import (
	"KeyMart/internal/utils"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"os"
	"strings"
)

func AddKey(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	adminID := os.Getenv("ADMIN_TELEGRAM_ID")
	if fmt.Sprintf("%d", update.Message.From.ID) != adminID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для выполнения этой команды.")
		bot.Send(msg)
		logger.Warn(fmt.Sprintf("Пользователь %d попытался использовать команду /addkey", update.Message.From.ID))
		return
	}

	args := update.Message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите ключ для добавления.")
		bot.Send(msg)
		return
	}

	_, err := db.Exec(context.Background(), "INSERT INTO vpn_keys (key_value) VALUES ($1)", args)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка добавления ключа: "+err.Error())
		bot.Send(msg)
		logger.Error(fmt.Sprintf("Ошибка добавления ключа: %v", err))
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ключ успешно добавлен!")
	bot.Send(msg)
	logger.Info(fmt.Sprintf("Администратор %d добавил ключ: %s", update.Message.From.ID, args))
}

func AllKeys(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	adminID := os.Getenv("ADMIN_TELEGRAM_ID") // Telegram ID администратора
	if fmt.Sprintf("%d", update.Message.From.ID) != adminID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администратору.")
		bot.Send(msg)
		return
	}

	rowsSold, err := db.Query(context.Background(),
		`SELECT u.username, u.telegram_id, 
                ARRAY_AGG(vk.key_value || ' (действителен до: ' || p.expires_at::TEXT || ')') AS keys
         FROM purchases p
         INNER JOIN vpn_keys vk ON p.key_id = vk.id
         INNER JOIN users u ON p.user_id = u.id
         GROUP BY u.username, u.telegram_id`)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка получения проданных ключей: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка проданных ключей.")
		bot.Send(msg)
		return
	}
	defer rowsSold.Close()

	var soldKeys []string
	for rowsSold.Next() {
		var username string
		var telegramID int64
		var userKeys []string

		err := rowsSold.Scan(&username, &telegramID, &userKeys)
		if err != nil {
			logger.Error(fmt.Sprintf("Ошибка сканирования строки для проданных ключей: %v", err))
			continue
		}

		var keysWithNumbers []string
		for i, key := range userKeys {
			keysWithNumbers = append(keysWithNumbers, fmt.Sprintf("%d. `%s`", i+1, key))
		}

		userInfo := fmt.Sprintf("*Пользователь:* %s (`%d`)", username, telegramID)
		keysInfo := fmt.Sprintf("*Ключи:*\n%s", strings.Join(keysWithNumbers, "\n"))

		soldKeys = append(soldKeys, fmt.Sprintf("%s\n%s", userInfo, keysInfo))
	}

	rowsAvailable, err := db.Query(context.Background(),
		"SELECT key_value FROM vpn_keys WHERE is_sold = FALSE")
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка получения доступных ключей: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка доступных ключей.")
		bot.Send(msg)
		return
	}
	defer rowsAvailable.Close()

	var availableKeys []string
	for rowsAvailable.Next() {
		var keyValue string
		err := rowsAvailable.Scan(&keyValue)
		if err != nil {
			logger.Error(fmt.Sprintf("Ошибка сканирования строки для доступных ключей: %v", err))
			continue
		}
		availableKeys = append(availableKeys, fmt.Sprintf("`%s`", keyValue))
	}

	var result strings.Builder
	if len(soldKeys) > 0 {
		result.WriteString("*Проданные ключи:*\n\n")
		result.WriteString(strings.Join(soldKeys, "\n\n"))
		result.WriteString("\n\n")
	} else {
		result.WriteString("*Проданные ключи:* Отсутствуют.\n\n")
	}

	if len(availableKeys) > 0 {
		result.WriteString("*Доступные ключи для продажи:*\n\n")
		result.WriteString(strings.Join(availableKeys, "\n"))
	} else {
		result.WriteString("*Доступные ключи для продажи:* Отсутствуют.")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, result.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)

	logger.Info("Список всех ключей отправлен администратору.")
}

func AddBulkKeys(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	adminID := os.Getenv("ADMIN_TELEGRAM_ID") // Telegram ID администратора
	if fmt.Sprintf("%d", update.Message.From.ID) != adminID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администратору.")
		bot.Send(msg)
		return
	}

	keysText := update.Message.CommandArguments()
	if keysText == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте ключи в сообщении, разделяя их строками.")
		bot.Send(msg)
		return
	}

	keys := strings.Split(keysText, "\n")
	var addedCount int

	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		_, err := db.Exec(context.Background(), "INSERT INTO vpn_keys (key_value) VALUES ($1)", key)
		if err != nil {
			logger.Error(fmt.Sprintf("Не удалось добавить ключ %s: %v", key, err))
			continue
		}
		addedCount++
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Успешно добавлено ключей: %d", addedCount))
	bot.Send(msg)
	logger.Info(fmt.Sprintf("Добавлено ключей: %d", addedCount))
}

func AddBulkKeysFromFile(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	adminID := os.Getenv("ADMIN_TELEGRAM_ID")
	if fmt.Sprintf("%d", update.Message.From.ID) != adminID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администратору.")
		bot.Send(msg)
		return
	}

	if update.Message.Document == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте текстовый файл с ключами.")
		bot.Send(msg)
		return
	}

	fileConfig := tgbotapi.FileConfig{FileID: update.Message.Document.FileID}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка получения файла: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось загрузить файл.")
		bot.Send(msg)
		return
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("TELEGRAM_BOT_TOKEN"), file.FilePath)
	response, err := http.Get(fileURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка скачивания файла: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка загрузки файла.")
		bot.Send(msg)
		return
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка чтения файла: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка чтения содержимого файла.")
		bot.Send(msg)
		return
	}

	keys := strings.Split(string(content), "\n")
	var addedCount int
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		_, err := db.Exec(context.Background(), "INSERT INTO vpn_keys (key_value) VALUES ($1)", key)
		if err != nil {
			logger.Error(fmt.Sprintf("Не удалось добавить ключ %s: %v", key, err))
			continue
		}
		addedCount++
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Успешно добавлено ключей: %d", addedCount))
	bot.Send(msg)
	logger.Info(fmt.Sprintf("Добавлено ключей из файла: %d", addedCount))
}

func HandlePreCheckoutQuery(query *tgbotapi.PreCheckoutQuery, bot *tgbotapi.BotAPI, logger *utils.Logger) {
	config := tgbotapi.PreCheckoutConfig{
		PreCheckoutQueryID: query.ID,
		OK:                 true,
	}
	_, err := bot.AnswerPreCheckoutQuery(config)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка подтверждения pre_checkout_query: %v", err))
	}
}

func HandleSuccessfulPayment(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *pgxpool.Pool, logger *utils.Logger) {
	userID := update.Message.From.ID
	payload := update.Message.SuccessfulPayment.InvoicePayload

	var keyID int
	_, err := fmt.Sscanf(payload, "user_%d_key_%d", &userID, &keyID)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка извлечения key_id из payload: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при обработке вашего платежа. Обратитесь к администратору.")
		bot.Send(msg)
		return
	}

	var status string
	err = db.QueryRow(context.Background(),
		"SELECT status FROM purchases WHERE user_id = (SELECT id FROM users WHERE telegram_id = $1) AND key_id = $2",
		userID, keyID).Scan(&status)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка проверки статуса платежа: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при проверке статуса вашего платежа.")
		bot.Send(msg)
		return
	}

	if status == "paid" {
		logger.Warn(fmt.Sprintf("Платеж для user_id %d и key_id %d уже обработан.", userID, keyID))
		return
	}

	var keyValue string
	err = db.QueryRow(context.Background(), "SELECT key_value FROM vpn_keys WHERE id = $1", keyID).Scan(&keyValue)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка извлечения ключа из базы данных: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении вашего ключа. Обратитесь к администратору.")
		bot.Send(msg)
		return
	}

	_, err = db.Exec(context.Background(),
		`UPDATE purchases 
         SET status = 'paid', expires_at = NOW() + INTERVAL '1 year' 
         WHERE user_id = (SELECT id FROM users WHERE telegram_id = $1) AND key_id = $2`, userID, keyID)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка обновления данных после оплаты: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при активации ключа. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	_, err = db.Exec(context.Background(),
		"UPDATE vpn_keys SET is_sold = TRUE WHERE id = $1", keyID)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка обновления статуса ключа: %v", err))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при обновлении статуса ключа. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Спасибо за оплату! Ваш ключ активирован:\n\n`%s`\n\nКлюч действителен до 1 года.", keyValue))
	msg.ParseMode = "Markdown"
	bot.Send(msg)

	logger.Info(fmt.Sprintf("Пользователь %d успешно оплатил ключ ID: %d. Ключ отправлен: %s", userID, keyID, keyValue))
}

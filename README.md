# KeyMart

KeyMart — это магазин для продажи VPN-ключей с использованием интеграции Telegram и платежной системы ЮKassa.

## Запуск проекта

### Требования:
- Go 1.20+
- PostgreSQL (если используется база данных)
- Docker (опционально для контейнеризации)

### Установка
1. Клонируйте репозиторий:
   ```bash
   git clone <repository_url>
   cd KeyMart
   ```

2. Создайте файл `.env` в корне проекта и заполните его переменные:
   ```env
   TELEGRAM_BOT_TOKEN=<ваш_токен>
   ADMIN_TELEGRAM_ID=<id_администратора>
   YU_KASSA_PROVIDER_TOKEN=<токен_юкассы>
   DATABASE_URL=<url_подключения_к_базе>
   SERVER_HOST=localhost
   SERVER_PORT=8080
   LOG_LEVEL=INFO
   JWT_SECRET=<секретный_ключ>
   ALLOWED_ORIGINS=*
   ENVIRONMENT=development
   ```

3. Установите зависимости:
   ```bash
   go mod tidy
   ```

4. Запустите проект:
   ```bash
   go run cmd/main.go
   ```

## Структура проекта

```
KeyMart/
├── .env                # Конфигурационные переменные среды
├── .gitignore          # Исключения для git
├── cmd/                # Точка входа в приложение
├── internal/           # Внутренние модули приложения
├── go.mod              # Зависимости Go
└── README.md           # Документация проекта
```

## Основной функционал
- Управление Telegram-ботом для взаимодействия с пользователями.
- Интеграция с платежной системой ЮKassa.
- Продажа VPN-ключей через Telegram.
- Панель администратора (опционально).

## Переменные окружения

| Переменная             | Описание                                    |
|------------------------|---------------------------------------------|
| `TELEGRAM_BOT_TOKEN`   | Токен Telegram-бота                        |
| `ADMIN_TELEGRAM_ID`    | ID администратора Telegram                 |
| `YU_KASSA_PROVIDER_TOKEN` | Токен для интеграции с ЮKassa            |
| `DATABASE_URL`         | URL подключения к базе данных              |


## Разработка

1. Запустите проект в режиме разработки:
   ```bash
   go run cmd/main.go
   ```

2. Запустите тесты:
   ```bash
   go test ./...
   ```

3. Соберите проект:
   ```bash
   go build -o keymart ./cmd
   ```

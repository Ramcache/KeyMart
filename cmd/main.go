package main

import (
	"KeyMart/internal/db"
	"KeyMart/internal/handlers"
	"KeyMart/internal/utils"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	logger := utils.NewLogger()

	dbConn := db.InitDB()
	defer dbConn.Close()

	handlers.StartBot(dbConn, logger)
}

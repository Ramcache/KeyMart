package utils

import (
	"log"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(message string) {
	log.Println("[INFO] " + message)
}

func (l *Logger) Error(message string) {
	log.Println("[ERROR] " + message)
}

func (l *Logger) Fatal(message string) {
	log.Fatal("[FATAL] " + message)
}

func (l *Logger) Panic(err error) {
	log.Panic("[PANIC] " + err.Error())
}

func (l *Logger) Warn(message string) {
	log.Println("[WARN] " + message)
}

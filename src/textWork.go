package main

import (
	"log"

	telego "github.com/mymmrac/telego"
)

func sendText(bot *telego.Bot, chatID int64, text string) {
	msg := &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	}
	if _, err := bot.SendMessage(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func sendFormattedText(bot *telego.Bot, chatID int64, text string) {
	msg := &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: chatID},
		Text:      text,
		ParseMode: "HTML",
	}
	if _, err := bot.SendMessage(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

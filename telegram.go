package main

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/telegram-bot-api.v4"
)

// TGBot is an internal struct to control telegram bot
type TGBot struct {
	api    tgbotapi.BotAPI
	chatID int64
}

func (t *TGBot) init() {
	var botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Panic("No BOT_TOKEN env var")
	}
	var botChatID = os.Getenv("BOT_CHATID")
	if botChatID == "" {
		log.Panic("No BOT_CHATID env var")
	}
	botChatIDInt, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		log.Panic("BOT_CHATID is not numeric")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	t.api = *bot
	t.chatID = botChatIDInt

	log.Printf("Authorized on account %s", bot.Self.UserName)
}

func (t *TGBot) sendMessage(message string) {
	log.Printf("[%d] %s", t.chatID, message)
	msg := tgbotapi.NewMessage(t.chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	t.api.Send(msg)
}

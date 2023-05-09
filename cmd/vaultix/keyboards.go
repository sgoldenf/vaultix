package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var restartKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Restart", "Restart"),
		tgbotapi.NewInlineKeyboardButtonData("Cancel", "Cancel"),
	),
)

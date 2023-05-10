package main

import (
	"flag"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	dbURL *string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("WARNING: No .env file found")
	}
	dbName := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbURL = flag.String(
		"dbURL",
		"postgres://"+user+":"+password+"@db:5432"+"/"+dbName,
		"PostgresSQL database URL",
	)
	flag.Parse()
}

func main() {
	app := newApplication()
	app.setHandlers()
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 25
	updates := app.bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message != nil && update.Message.Command() != "" {
			command := update.Message.Command()
			handler, ok := app.messageHandlers[command]
			if !ok {
				handler = app.messageHandlers["default"]
			}
			if err := handler(update.Message); err != nil {
				app.errorLog.Println(err)
				app.errorNotification(update.Message.From.ID)
			}
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(
				update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := app.bot.Request(callback); err != nil {
				app.errorLog.Println(err)
				continue
			}
			app.infoLog.Println(update.CallbackQuery.Data)
			if update.CallbackQuery.Data == "Restart" {
				if err := app.deleteUser(update.CallbackQuery.From.ID); err != nil {
					app.errorLog.Println(err)
					app.errorNotification(update.CallbackQuery.From.ID)
				} else if err = app.createUser(update.CallbackQuery.From.ID); err != nil {
					app.errorLog.Println(err)
					app.errorNotification(update.CallbackQuery.From.ID)
				}
			} else if update.CallbackQuery.Data == "Cancel" {
				deleteMsg := tgbotapi.NewDeleteMessage(
					update.FromChat().ChatConfig().ChatID,
					update.CallbackQuery.Message.MessageID,
				)
				if _, err := app.bot.Send(deleteMsg); err != nil {
					app.errorLog.Println(err)
				}
			}
		}
	}
}

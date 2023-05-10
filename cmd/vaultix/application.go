package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sgoldenf/vaultix/internal/adapters/message"
	"github.com/sgoldenf/vaultix/internal/adapters/password"
	"github.com/sgoldenf/vaultix/internal/adapters/user"
	"github.com/sgoldenf/vaultix/internal/models"
)

type application struct {
	bot             *tgbotapi.BotAPI
	userModel       user.UserModelInterface
	passwordModel   password.PasswordModelInterface
	messageModel    message.MessageModelInterface
	infoLog         *log.Logger
	errorLog        *log.Logger
	messageHandlers map[string]func(*tgbotapi.Message) error
}

func newApplication() *application {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	if err := godotenv.Load(); err != nil {
		errorLog.Fatal("no .env file found")
	}
	db, err := dbConn()
	if err != nil {
		errorLog.Fatal(err)
	}
	apiToken := os.Getenv("BOT_API_TOKEN")
	infoLog.Println("authorization using bot api token: " + apiToken)
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Printf("Authorized on account %s", bot.Self.UserName)
	return &application{
		bot:           bot,
		userModel:     &models.UserModel{Pool: db},
		passwordModel: &models.PasswordModel{Pool: db},
		messageModel:  &models.MessageModel{Pool: db},
		infoLog:       infoLog,
		errorLog:      errorLog,
	}
}

func dbConn() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), *dbURL)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return conn, err
}

func (app *application) errorNotification(userID int64) {
	app.notifyUser(userID,
		`Oops, something went wrong\. Please try again later or contact @sgoldenf for technical support\.`,
		tgbotapi.ModeMarkdownV2,
	)
}

func (app *application) notifyUser(userID int64, text string, parseMode string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = parseMode
	if message, err := app.bot.Send(msg); err != nil {
		app.errorLog.Println(err)
	} else {
		app.messageModel.AddMessage(message.Chat.ChatConfig().ChatID, message.MessageID)
	}
}

func (app *application) createUser(userID int64) error {
	masterPassword, err := app.userModel.CreateUser(userID)
	if err != nil {
		return err
	}
	app.notifyUser(userID,
		`Here is your Master Password: `+"`"+masterPassword+"`"+`
You'll need it to have access to saved passwords\.
Store it in a safe place and then you can delete this message\.
WARNING: If you'll loose/forget your password, you won't have access to your data\.
You'll be able to restart bot using /start command and get new Master Password, but all previous data will be lost\.`,
		tgbotapi.ModeMarkdownV2,
	)
	return nil
}

func (app *application) deleteUser(UserID int64) error {
	delPasswords, delUsers, err := app.userModel.DeleteUser(UserID)
	if err != nil {
		return err
	}
	app.notifyUser(UserID,
		fmt.Sprint("Deleted ", delPasswords, " passwords from ", delUsers, " user(s)."), "")
	return nil
}

func (app *application) deleteExpiredMessagesFromBotRoutine() {
	for {
		messages, err := app.messageModel.DeleteExpired()
		if err != nil {
			app.errorLog.Println(err)
		}
		for _, message := range messages {
			app.deleteMessage(message.ChatID, message.MessageID)
		}
		time.Sleep(5 * time.Minute)
	}
}

func (app *application) deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(
		chatID,
		messageID,
	)
	if _, err := app.bot.Request(deleteMsg); err != nil {
		app.errorLog.Println(err)
	}
}

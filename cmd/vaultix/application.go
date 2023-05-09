package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sgoldenf/vaultix/internal/adapters/password"
	"github.com/sgoldenf/vaultix/internal/adapters/user"
	"github.com/sgoldenf/vaultix/internal/models"
)

type application struct {
	bot             *tgbotapi.BotAPI
	userModel       user.UserModelInterface
	passwordModel   password.PasswordModelInterface
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
	bot.Debug = true
	infoLog.Printf("Authorized on account %s", bot.Self.UserName)
	return &application{
		bot:           bot,
		userModel:     &models.UserModel{Pool: db},
		passwordModel: &models.PasswordModel{Pool: db},
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

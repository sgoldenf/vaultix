package main

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sgoldenf/vaultix/internal/models"
)

func (app *application) setHandlers() {
	app.messageHandlers = map[string]func(*tgbotapi.Message) error{
		"start":   app.handleStart,
		"set":     app.handleSet,
		"get":     app.handleGet,
		"del":     app.handleDel,
		"help":    app.handleHelp,
		"default": app.handleDefault,
	}
}

func (app *application) handleStart(message *tgbotapi.Message) error {
	if exists, err := app.userModel.Exists(message.From.ID); err != nil {
		return err
	} else if exists {
		app.handleExistingUser(message.From.ID)
		return nil
	}
	if err := app.createUser(message.From.ID); err != nil {
		return err
	}
	return nil
}

func (app *application) handleExistingUser(userID int64) {
	msg := tgbotapi.NewMessage(userID,
		`User with your Telegram ID already exists.
If you forgot your Master Password and want to get a new one, press Restart Button.
WARNING: if you proceed, all data for your account will be deleted`)
	msg.ReplyMarkup = restartKeyboard
	if _, err := app.bot.Send(msg); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) handleSet(message *tgbotapi.Message) error {
	args := strings.Split(message.CommandArguments(), " ")
	if len(args) != 4 {
		app.notifyUser(message.From.ID,
			`Usage: "/set <service> <login> <password> <Master Password>"`, "")
		return nil
	}
	if _, err := app.userModel.Authenticate(message.From.ID, args[3]); err != nil {
		if err == models.ErrInvalidMasterPassword {
			app.notifyUser(message.From.ID,
				`Invalid Master Password`, "")
			return nil
		} else {
			return err
		}
	}
	if err := app.passwordModel.AddPassword(
		message.From.ID, args[0], args[1], args[2], args[3],
	); err != nil {
		if err == models.ErrDuplicateCredentials {
			app.notifyUser(message.From.ID,
				`There already is a password for this service-login pair.
You can use /del to delete it first.`, "")
			return nil
		} else {
			return err
		}
	}
	app.notifyUser(message.From.ID,
		`Your credentials to this service has been successfuly added to password manager.
For security purposes, we recommend you to delete your messages containing any credentials.`, "")
	return nil
}

func (app *application) handleGet(message *tgbotapi.Message) error {
	args := strings.Split(message.CommandArguments(), " ")
	if len(args) != 2 {
		app.notifyUser(message.From.ID, `Usage: "/get <service> <Master Password>"`, "")
		return nil
	}
	if _, err := app.userModel.Authenticate(message.From.ID, args[1]); err != nil {
		if err == models.ErrInvalidMasterPassword {
			app.notifyUser(message.From.ID, `Invalid Master Password`, "")
			return nil
		} else {
			return err
		}
	}
	passwords, err := app.passwordModel.GetPasswords(message.From.ID, args[0], args[1])
	if err != nil {
		return err
	}
	if len(passwords) == 0 {
		app.notifyUser(message.From.ID,
			"You don't have credentials for "+args[0]+" service", "")
		return nil
	}
	text := "Your credentials for " + args[0] + " service:"
	for _, password := range passwords {
		text += fmt.Sprintf("\n\nLogin: `%s`\nPassword: `%s`", password.Login, password.Password)
	}
	app.notifyUser(message.From.ID, text, tgbotapi.ModeMarkdownV2)
	return nil
}

func (app *application) handleDel(message *tgbotapi.Message) error {
	return nil
}

func (app *application) handleHelp(message *tgbotapi.Message) error {
	return nil
}

func (app *application) handleDefault(message *tgbotapi.Message) error {
	return nil
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
	if _, err := app.bot.Send(msg); err != nil {
		app.errorLog.Println(err)
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

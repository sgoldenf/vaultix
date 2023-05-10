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
	defer app.deleteMessage(message.Chat.ID, message.MessageID)
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
	if exists, err := app.passwordModel.Exists(message.From.ID, args[0], args[1]); err != nil {
		return err
	} else if exists {
		app.notifyUser(message.From.ID,
			`There already is a password for this service-login pair.
You can use /del to delete it first.`, "")
		return nil
	}
	if err := app.passwordModel.AddPassword(
		message.From.ID, args[0], args[1], args[2], args[3],
	); err != nil {
		return err
	}
	app.notifyUser(message.From.ID,
		`Your credentials to this service has been successfuly added to password manager.`, "")
	return nil
}

func (app *application) handleGet(message *tgbotapi.Message) error {
	defer app.deleteMessage(message.Chat.ID, message.MessageID)
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
	args := strings.Split(message.CommandArguments(), " ")
	if len(args) != 1 {
		app.notifyUser(message.From.ID,
			`Usage: "/del <service>"`, "",
		)
		return nil
	}
	if deleted, err := app.passwordModel.DeletePasswords(message.From.ID, args[0]); err != nil {
		return err
	} else if deleted == 0 {
		app.notifyUser(message.From.ID,
			`You don't have credentials for `+args[0]+` service.`, "",
		)
	} else {
		app.notifyUser(message.From.ID,
			fmt.Sprint(deleted, ` passwords deleted from `, args[0], ` service.`), "",
		)
	}
	return nil
}

func (app *application) handleHelp(message *tgbotapi.Message) error {
	app.notifyUser(message.From.ID,
		`Usage:
/start - init using vaultix and get Master Password
/set <service> <login> <password> <Master Password> - add login and password to service
/get <service> <Master Password> - retrieve login and password by service name
/del <service> <login> - delete login and password for service
/help all available commands`, "",
	)
	return nil
}

func (app *application) handleDefault(message *tgbotapi.Message) error {
	app.notifyUser(message.From.ID,
		`I don't know `+message.Command()+` command :(`, "",
	)
	return nil
}

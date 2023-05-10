package message

import "github.com/sgoldenf/vaultix/internal/models"

type MessageModelInterface interface {
	AddMessage(chatID int64, messageID int) error
	DeleteExpired() ([]*models.Message, error)
}

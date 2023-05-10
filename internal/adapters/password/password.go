package password

import "github.com/sgoldenf/vaultix/internal/models"

type PasswordModelInterface interface {
	AddPassword(userID int64, service, login, password, masterPassword string) error
	GetPasswords(userID int64, service, masterPassword string) ([]*models.Password, error)
}

package password

type PasswordModelInterface interface {
	AddPassword(userID int64, service, login, password, masterPassword string) error
}

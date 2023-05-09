package user

type UserModelInterface interface {
	CreateUser(userID int64) (string, error)
	Authenticate(userID int64, password string) (int, error)
	Exists(userID int64) (bool, error)
	DeleteUser(userID int64) (passwordsDeleted int64, usersDeleted int64, err error)
}

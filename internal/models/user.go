package models

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sgoldenf/vaultix/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// type User struct {
// 	TelegramID     int64
// 	HashedPassword []byte
// }

type UserModel struct {
	Pool *pgxpool.Pool
}

func (m *UserModel) CreateUser(userID int64) (string, error) {
	key, err := utils.GenerateKey()
	if err != nil {
		return "", err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(key), 12)
	if err != nil {
		return "", err
	}
	_, err = m.Pool.Query(context.Background(),
		`insert into users (id, hashed_password) values ($1, $2);`,
		userID, string(hashedPassword))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Check if error code is unique_violation (42710)
			if pgErr.Code == "23505" {
				return "", ErrDuplicateID
			}
		}
		return "", err
	}
	return key, nil
}

func (m *UserModel) Authenticate(userID int64, password string) (int, error) {
	var id int
	var hashedPassword []byte
	err := m.Pool.QueryRow(context.Background(),
		`select id, hashed_password from users where id = $1;`,
		userID).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidMasterPassword
		} else {
			return 0, err
		}
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidMasterPassword
		} else {
			return 0, err
		}
	}
	return id, nil
}

func (m *UserModel) Exists(userID int64) (bool, error) {
	var exists bool
	err := m.Pool.QueryRow(context.Background(),
		`select exists(select true from users where id = $1);`, userID).Scan(&exists)
	return exists, err
}

func (m *UserModel) DeleteUser(userID int64) (passwordsDeleted int64, usersDeleted int64, err error) {
	tx, err := m.Pool.Begin(context.Background())
	if err != nil {
		return
	}
	defer tx.Rollback(context.Background())
	res1, err := tx.Exec(
		context.Background(),
		`delete from passwords where user_id = $1;`, userID)
	if err != nil {
		return
	}
	res2, err := tx.Exec(
		context.Background(),
		`delete from users where id = $1;`, userID)
	if err != nil {
		return
	}
	passwordsDeleted = res1.RowsAffected()
	usersDeleted = res2.RowsAffected()
	if err = tx.Commit(context.Background()); err != nil {
		return
	}
	return
}

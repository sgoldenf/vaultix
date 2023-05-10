package models

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sgoldenf/vaultix/internal/utils"
)

type Password struct {
	Login    string
	Password string
}

type PasswordModel struct {
	Pool *pgxpool.Pool
}

func (m *PasswordModel) AddPassword(userID int64, service, login, password, masterPassword string) error {
	encryptedPassword, err := utils.EncryptPassword([]byte(password), []byte(masterPassword))
	if err != nil {
		return err
	}
	_, err = m.Pool.Query(context.Background(),
		`insert into passwords (user_id, service, login, encrypted_password) values ($1, $2, $3, $4) returning id;`,
		userID, service, login, encryptedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Check if error code is duplicate_object (42710)
			// Check if error message contains passwords_user_service_login_uc (constraint for unique credentials)
			if pgErr.Code == "42710" && strings.Contains(pgErr.Message, "passwords_user_service_login_uc") {
				return ErrDuplicateCredentials
			}
		}
		return err
	}
	return nil
}

func (m *PasswordModel) GetPasswords(userID int64, service, masterPassword string) ([]*Password, error) {
	rows, err := m.Pool.Query(context.Background(),
		`select login, encrypted_password from passwords where user_id=$1 and service = $2;`,
		userID, service,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var passwords []*Password
	for rows.Next() {
		var login string
		var encrypted_password []byte
		err = rows.Scan(&login, &encrypted_password)
		if err != nil {
			return nil, err
		}
		password, err := utils.DecryptPassword(encrypted_password, []byte(masterPassword))
		if err != nil {
			return nil, err
		}
		passwords = append(passwords, &Password{
			Login:    login,
			Password: password,
		})
	}
	return passwords, nil
}

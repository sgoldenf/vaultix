package models

import (
	"context"

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
	encryptedPassword, err := utils.EncryptPassword(password, masterPassword)
	if err != nil {
		return err
	}
	rows, err := m.Pool.Query(context.Background(),
		`insert into passwords (user_id, service, login, encrypted_password) values ($1, $2, $3, $4) returning id;`,
		userID, service, login, encryptedPassword)
	if err != nil {
		return err
	}
	defer rows.Close()
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
		var encrypted_password string
		if err := rows.Scan(&login, &encrypted_password); err != nil {
			return nil, err
		}
		if password, err := utils.DecryptPassword(encrypted_password, masterPassword); err != nil {
			return nil, err
		} else {
			passwords = append(passwords, &Password{
				Login:    login,
				Password: password,
			})
		}
	}
	return passwords, nil
}

func (m *PasswordModel) DeletePasswords(userID int64, service string) (int64, error) {
	res, err := m.Pool.Exec(context.Background(),
		`delete from passwords where user_id = $1 and service = $2;`,
		userID, service,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (m *PasswordModel) Exists(userID int64, service, login string) (bool, error) {
	var exists bool
	err := m.Pool.QueryRow(context.Background(),
		`select exists(select true from passwords where user_id = $1 and service = $2 and login = $3);`,
		userID, service, login,
	).Scan(&exists)
	return exists, err
}

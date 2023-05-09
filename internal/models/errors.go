package models

import "errors"

var (
	ErrNoRecord              = errors.New("no matching record found")
	ErrInvalidMasterPassword = errors.New("invalid master password")
	ErrDuplicateID           = errors.New("user with this telegram id already exists")
	ErrDuplicateCredentials  = errors.New("there already is a password for this service-login pair")
)

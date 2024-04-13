package models

import "errors"

var (
	ErrInvalidAccountAccessOption = errors.New("invalid account access option")
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrUserNotExists              = errors.New("user not exists")
	ErrInvalidOperation           = errors.New("invalid operation")
	ErrUserDoNotHaveGithubAccess  = errors.New("user do not have github access")
	ErrInvalidEmailAddr           = errors.New("invalid email address")
	ErrWrongVerificationCode      = errors.New("wrong verification code")
)

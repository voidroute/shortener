package repository

import "errors"

var (
	ErrLinkNotFound = errors.New("link not found")
	ErrCodeExists   = errors.New("code already exists")
)

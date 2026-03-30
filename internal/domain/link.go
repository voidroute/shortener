package domain

import (
	"fmt"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	_ = validate.RegisterValidation("code", func(fl validator.FieldLevel) bool {
		for _, c := range fl.Field().String() {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
				return false
			}
		}
		return true
	})
}

type Link struct {
	URL       string  `validate:"required,url"`
	Code      string  `validate:"required,code,min=3,max=16"`
	Alias     *string `validate:"omitempty,code,min=3,max=16"`
	ExpiresAt *time.Time
	CreatedAt time.Time
}

func (l *Link) Validate() error {
	if err := validate.Struct(l); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidLink, err.Error())
	}

	return nil
}

func NewLink(url, code string, alias *string, expiresAt *time.Time) (*Link, error) {
	link := &Link{
		URL:       url,
		Code:      code,
		Alias:     alias,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := link.Validate(); err != nil {
		return nil, err
	}

	return link, nil
}

func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*l.ExpiresAt)
}

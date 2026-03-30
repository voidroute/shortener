package domain

import "errors"

var ErrInvalidLink = errors.New("invalid link")
var ErrLinkExpired = errors.New("link expired")

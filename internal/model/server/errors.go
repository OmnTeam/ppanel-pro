package server

import "errors"

var (
	ErrProtocolTypeRequired  = errors.New("protocol type is required")
	ErrDuplicateProtocolType = errors.New("duplicate protocol type")
	ErrInvalidProtocolConfig = errors.New("invalid protocol configuration")
)

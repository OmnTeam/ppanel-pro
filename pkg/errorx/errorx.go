package errorx

import (
	"errors"
)

// New creates a new error
func New(text string) error {
	return errors.New(text)
}

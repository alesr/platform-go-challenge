package users

import (
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	// Enumerate service errors

	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUserValue = errors.New("could not cast store value to user.User")
)

type User struct {
	ID        ulid.ULID
	Name      string `faker:"name,unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

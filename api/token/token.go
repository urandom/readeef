package token

import "time"

type Storage interface {
	Store(token string, expiration time.Time) error
	Exists(token string) (bool, error)
	RemoveExpired() error
}

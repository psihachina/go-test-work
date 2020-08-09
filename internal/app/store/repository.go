package store

import "github.com/psihachina/go-test-work.git/internal/app/model"

type TokenRepository interface {
	CreateAuth(string, *model.TokenDetails) error
	DeleteToken(*model.AccessDetails) error
	DeleteTokens(*model.AccessDetails) error
	DeleteAuth(string) (int64, error)
}

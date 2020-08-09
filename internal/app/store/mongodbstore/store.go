package mongodbstore

import (
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"go.mongodb.org/mongo-driver/mongo"
)

// Store ...
type Store struct {
	db              *mongo.Database
	tokenRepository *TokenRepository
}

// New ...
func New(db *mongo.Database) *Store {
	return &Store{
		db: db,
	}
}

// Token ...
func (s *Store) Token() store.TokenRepository {
	if s.tokenRepository != nil {
		return s.tokenRepository
	}

	s.tokenRepository = &TokenRepository{
		store: s,
	}

	return s.tokenRepository
}

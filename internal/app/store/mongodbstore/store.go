package mongodbstore

import (
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"go.mongodb.org/mongo-driver/mongo"
)

// Store ...
type Store struct {
	db             *mongo.Database
	userRepository *UserRepository
}

// New ...
func New(db *mongo.Database) *Store {
	return &Store{
		db: db,
	}
}

// User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}
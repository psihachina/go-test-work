package store

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// Store ...
type Store struct {
	config         *Config
	db             *mongo.Database
	userRepository *UserRepository
}

// New ...
func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

// Open ...
func (s *Store) Open(database string) error {
	db, err := mongo.NewClient(options.Client().ApplyURI(s.config.DatabaseURL))
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = db.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	s.db = db.Database(database)

	return nil
}

// Close ...
func (s *Store) Close() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err := s.db.Client().Disconnect(ctx); err != nil {
		log.Fatal(err)
	}
}

// User ...
func (s *Store) User() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

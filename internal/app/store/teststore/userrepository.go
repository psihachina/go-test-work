package teststore

import (
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository ...
type UserRepository struct {
	store *Store
	users map[string]*model.User
}

// Create ...
func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	r.users[u.Email] = u
	u.ID = primitive.NewObjectID()

	return nil
}

// FindByEmail ...
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u, ok := r.users[email]

	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}

type Fields struct {
	ID       primitive.ObjectID
	Email    string
	Password string
}

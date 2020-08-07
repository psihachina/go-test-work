package mongodbstore

import (
	"context"
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type UserRepository struct {
	store *Store
}

// Create ...
func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	userCollection := r.store.db.Collection("users")

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if res, err := userCollection.InsertOne(sessCtx, bson.D{{"email", u.Email}, {"encrypted_password", u.EncryptedPassword}}); err != nil {
			return nil, err
		} else {
			u.ID = res.InsertedID.(primitive.ObjectID)
		}

		return u, nil
	}

	session, err := r.store.db.Client().StartSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.EndSession(context.Background())

	_, err = session.WithTransaction(context.Background(), callback)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

// FindByEmail ...
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	var result Fields
	if err := r.store.db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}
	u.Email = result.Email
	u.ID = result.ID
	u.EncryptedPassword = result.Password

	return u, nil
}

type Fields struct {
	ID       primitive.ObjectID
	Email    string
	Password string
}

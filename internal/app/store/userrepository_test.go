package store_test

import (
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	s, teardown := store.TestStore(t, databaseUrl)
	defer teardown("users")

	u, err := s.User().Create(&model.User{
		Email:             "test@test@gmail.com",
		EncryptedPassword: "test1234",
	})

	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	s, teardown := store.TestStore(t, databaseUrl)
	defer teardown("users")

	_, err := s.User().FindByEmail("test@test@gmail.com")
	assert.Error(t, err)

	s.User().Create(&model.User{
		Email:             "test@test@gmail.com",
		EncryptedPassword: "test1234",
	})

	u, err := s.User().FindByEmail("test@test@gmail.com")

	assert.NoError(t, err)
	assert.NotNil(t, u)
}

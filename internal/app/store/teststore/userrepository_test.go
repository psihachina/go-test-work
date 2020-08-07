package teststore_test

import (
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"github.com/psihachina/go-test-work.git/internal/app/store/teststore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	s := teststore.New()

	u := model.TestUser(t)

	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u)
}

func TestUserRepository_FindByEmail(t *testing.T) {

	s := teststore.New()

	email := "test@gmail.com"

	_, err := s.User().FindByEmail(email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser(t)
	u.Email = email

	_ = s.User().Create(u)

	u, err = s.User().FindByEmail("test@gmail.com")

	assert.NoError(t, err)
	assert.NotNil(t, u)
}

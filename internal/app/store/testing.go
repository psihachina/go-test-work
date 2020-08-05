package store

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

// TestStore ...
func TestStore(t *testing.T, databaseUrl string) (*Store, func(...string)) {
	t.Helper()

	config := NewConfig()
	config.DatabaseURL = databaseUrl
	s := New(config)
	if err := s.Open("test_database"); err != nil {
		t.Fatal(err)
	}

	return s, func(tables ...string) {
		if len(tables) > 0 {
			for _, value := range tables {
				if _, err := s.db.Collection(value).DeleteMany(context.Background(), bson.M{}); err != nil {
					t.Fatal(err)
				}
			}
			s.Close()
		}
	}
}

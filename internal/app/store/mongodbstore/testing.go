package mongodbstore

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"testing"
	"time"
)

// TestStore ...
func TestDB(t *testing.T, databaseUrl string, databaseName string) (*mongo.Database, func(...string)) {
	t.Helper()

	db, err := mongo.NewClient(options.Client().ApplyURI(databaseUrl))
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = db.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return db.Database(databaseName), func(tables ...string) {
		if len(tables) > 0 {
			for _, value := range tables {
				if _, err := db.Database(databaseName).Collection(value).DeleteMany(context.Background(), bson.M{}); err != nil {
					t.Fatal(err)
				}
			}
			db.Disconnect(ctx)
		}
	}
}

package apiserver

import (
	"context"
	"github.com/psihachina/go-test-work.git/internal/app/store/mongodbstore"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseUrl)
	if err != nil {
		return err
	}

	defer db.Client().Disconnect(context.Background())

	store := mongodbstore.New(db)

	srv := newServer(store)

	return http.ListenAndServe(config.BindAddr, srv)
}

func newDB(databaseUrl string) (*mongo.Database, error) {
	db, err := mongo.NewClient(options.Client().ApplyURI(databaseUrl))
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = db.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return db.Database("test_work"), nil
}

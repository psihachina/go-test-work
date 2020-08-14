package mongodbstore

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/psihachina/go-test-work.git/internal/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TokenRepository struct {
	store *Store
}

func (r *TokenRepository) CreateAuth(userid string, td *model.TokenDetails) error {
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	var ctx = context.Background()
	var err error
	var session mongo.Session
	var resultRef *mongo.InsertOneResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("refresh_sessions")
		if resultRef, err = collection.InsertOne(sc, bson.M{
			"refreshToken": td.RefreshUuid,
			"userId":       userid,
			"createAt":     rt.Sub(now)}); err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		if resultRef.InsertedID == nil {
			log.Fatal("insert failed, expected id but got ", resultRef.InsertedID)
		}

		if err = session.CommitTransaction(sc); err != nil {
			log.Fatal(err)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	session.EndSession(ctx)

	return nil
}

func (r *TokenRepository) DeleteTokens(authD *model.AccessDetails) error {
	var ctx = context.Background()
	var err error
	var session mongo.Session
	var deletedRt *mongo.DeleteResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("refresh_sessions")
		deletedRt, err = collection.DeleteMany(context.Background(), bson.M{"userId": authD.UserID})
		if err != nil {
			log.Fatal(err)
		}
		if err = session.CommitTransaction(sc); err != nil {
			log.Fatal(err)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	session.EndSession(ctx)

	return nil
}

func (r *TokenRepository) DeleteAuth(givenUuid string) (int64, error) {
	var ctx = context.Background()
	var err error
	var session mongo.Session
	var deletedRt *mongo.DeleteResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("refresh_sessions")
		deletedRt, err = collection.DeleteOne(context.Background(), bson.M{"refreshToken": givenUuid})
		if err != nil {
			return nil
		}

		if deletedRt.DeletedCount != 1 {
			fmt.Println("Token not found")
			return session.AbortTransaction(sc)
		}

		if err = session.CommitTransaction(sc); err != nil {
			log.Fatal(err)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	session.EndSession(ctx)

	return deletedRt.DeletedCount, nil
}

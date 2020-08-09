package mongodbstore

import (
	"context"
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type TokenRepository struct {
	store *Store
}

func (r *TokenRepository) CreateAuth(userid string, td *model.TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	var ctx = context.Background()
	var err error
	var session mongo.Session
	var resultAcc *mongo.InsertOneResult
	var resultRef *mongo.InsertOneResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("access_sessions")
		if resultAcc, err = collection.InsertOne(sc, bson.M{
			"accessToken": td.AccessUuid,
			"userId":      userid,
			"createAt":    at.Sub(now)}); err != nil {
			log.Fatal(err)
		}
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
		if resultAcc.InsertedID == nil || resultRef.InsertedID == nil {
			log.Fatal("insert failed, expected id but got ", resultAcc.InsertedID, resultRef.InsertedID)
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

func (r *TokenRepository) DeleteToken(authD *model.AccessDetails) error {
	var ctx = context.Background()
	var err error
	var session mongo.Session
	var deletedAt *mongo.DeleteResult
	var deletedRt *mongo.DeleteResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("access_sessions")
		deletedAt, err = collection.DeleteOne(context.Background(), bson.M{"accessToken": authD.AccessUuid})
		if err != nil {
			log.Fatal(err)
		}
		collection = r.store.db.Collection("refresh_sessions")
		deletedRt, err = collection.DeleteOne(context.Background(), bson.M{"refreshToken": authD.RefreshUuid})
		if err != nil {
			log.Fatal(err)
		}

		if deletedAt.DeletedCount != 1 || deletedRt.DeletedCount != 1 {
			log.Fatal("delete failed, expected count but got ", deletedAt.DeletedCount, deletedRt.DeletedCount)
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
	var deletedAt *mongo.DeleteResult
	var deletedRt *mongo.DeleteResult
	var collection *mongo.Collection

	if session, err = r.store.db.Client().StartSession(); err != nil {
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("access_sessions")
		deletedAt, err = collection.DeleteMany(context.Background(), bson.M{"userId": authD.UserId})
		if err != nil {
			log.Fatal(err)
		}
		collection = r.store.db.Collection("refresh_sessions")
		deletedRt, err = collection.DeleteMany(context.Background(), bson.M{"userId": authD.UserId})
		if err != nil {
			log.Fatal(err)
		}

		if deletedAt.DeletedCount == 0 || deletedRt.DeletedCount == 0 {
			log.Fatal("delete failed, expected count but got ", deletedAt.DeletedCount, deletedRt.DeletedCount)
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
		log.Fatal(err)
	}
	if err = session.StartTransaction(); err != nil {
		log.Fatal(err)
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		collection = r.store.db.Collection("refresh_sessions")
		deletedRt, err = collection.DeleteOne(context.Background(), bson.M{"refreshToken": givenUuid})
		if err != nil {
			log.Fatal(err)
		}

		if deletedRt.DeletedCount != 1 {
			log.Fatal("delete failed, expected count but got ", deletedRt.DeletedCount)
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

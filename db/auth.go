package db

import (
	"fractapp-server/notification"

	"go.mongodb.org/mongo-driver/bson"
)

type Auth struct {
	Id        ID                           `bson:"_id"`
	Value     string                       `bson:"value"`
	IsValid   bool                         `bson:"is_valid"`
	Code      string                       `bson:"code"`
	Attempts  int32                        `bson:"attempts"`
	Count     int32                        `bson:"count"`
	Timestamp int64                        `bson:"timestamp"`
	Type      notification.NotificatorType `bson:"type"`
}

func (db *MongoDB) AuthByValue(value string, codeType notification.NotificatorType) (*Auth, error) {
	collection := db.collections[AuthDB]
	auth := &Auth{}
	res := collection.FindOne(db.ctx, bson.D{
		{"value", value},
		{"type", codeType},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(auth)
	if err != nil {
		return nil, err
	}
	return auth, nil
}

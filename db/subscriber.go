package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Subscriber struct {
	Id        ID     `bson:"_id"`
	ProfileId ID     `bson:"profile"`
	Token     string `bson:"token"`
	Timestamp int64  `bson:"timestamp"`
}

func (db *MongoDB) SubscribersCountByToken(token string) (int64, error) {
	collection := db.collections[SubscribersDB]

	count, err := collection.CountDocuments(db.ctx, bson.D{
		{"token", token},
	}, nil)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *MongoDB) SubscriberByProfileId(id ID) (*Subscriber, error) {
	var subscriber *Subscriber

	collection := db.collections[SubscribersDB]

	res := collection.FindOne(db.ctx, bson.D{
		{"profile", id},
	}, nil)
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(subscriber)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

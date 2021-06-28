package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	Id          ID     `bson:"_id"`
	Value       string `bson:"value"`
	SenderId    ID     `bson:"sender_id"`   //TODO ref
	ReceiverId  ID     `bson:"receiver_id"` //TODO ref
	Timestamp   int64  `bson:"timestamp"`
	IsDelivered bool   `bson:"is_delivered"`
}

func (db *MongoDB) MessagesByReceiver(receiver ID) ([]Message, error) {
	collection := db.collections[MessagesDB]

	opt := options.FindOne()
	opt.SetSort(bson.D{{"timestamp", 1}})

	var messages []Message
	res := collection.FindOne(db.ctx, bson.D{
		{"receiver_id", receiver},
		{"is_delivered", false},
	}, opt)
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *MongoDB) MessagesBySenderAndReceiver(sender ID, receiver ID) ([]Message, error) {
	collection := db.collections[MessagesDB]

	opt := options.FindOne()
	opt.SetSort(bson.D{{"timestamp", 1}})

	var messages []Message
	res := collection.FindOne(db.ctx, bson.D{
		{"sender_id", sender},
		{"receiver_id", receiver},
		{"is_delivered", false},
	}, opt)
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *MongoDB) UpdateDeliveredMessage(id primitive.ObjectID) error {
	collection := db.collections[MessagesDB]

	_, err := collection.UpdateByID(db.ctx, id, nil)
	if err != nil {
		return err
	}

	return nil
}

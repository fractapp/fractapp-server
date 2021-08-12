package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Row struct {
	Buttons []Button `json:"buttons" bson:"buttons"`
}

type Button struct {
	Value     string   `json:"value" bson:"value"`
	Action    string   `json:"action" bson:"action"`
	Arguments []string `json:"arguments" bson:"arguments"`
	ImageUrl  string   `json:"imageUrl" bson:"image_url"`
}

type Message struct {
	Id      ID       `bson:"_id"`
	Version int      `bson:"version"`
	Action  string   `bson:"action"`
	Value   string   `bson:"value"`
	Args    []string `bson:"args"`
	Rows    []Row    `bson:"rows"`

	SenderId    ID    `bson:"sender_id"`   //TODO ref
	ReceiverId  ID    `bson:"receiver_id"` //TODO ref
	Timestamp   int64 `bson:"timestamp"`
	IsDelivered bool  `bson:"is_delivered"`
}

func (db *MongoDB) MessagesByReceiver(receiver ID) ([]Message, error) {
	collection := db.collections[MessagesDB]

	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	messages := make([]Message, 0)

	res, err := collection.Find(db.ctx, bson.D{
		{"receiver_id", receiver},
		{"is_delivered", false},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *MongoDB) SetDelivered(owner ID, id ID) error {
	collection := db.collections[MessagesDB]

	_, err := collection.UpdateOne(db.ctx, bson.D{
		{"_id", id},
		{"receiver_id", owner},
	}, bson.D{
		{"$set", bson.D{{"is_delivered", true}}},
	})
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoDB) MessagesBySenderAndReceiver(sender ID, receiver ID) ([]Message, error) {
	collection := db.collections[MessagesDB]

	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	messages := make([]Message, 0)
	res, err := collection.Find(db.ctx, bson.D{
		{"sender_id", sender},
		{"receiver_id", receiver},
		{"is_delivered", false},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

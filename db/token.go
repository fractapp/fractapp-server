package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Token struct {
	Id        ID     `bson:"_id"`
	ProfileId ID     `bson:"profile"`
	Token     string `bson:"token"`
}

func (db *MongoDB) TokenByValue(token string) (*Token, error) {
	tokenDb := &Token{}

	collection := db.collections[TokensDB]

	res := collection.FindOne(db.ctx, bson.D{
		{"token", token},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(tokenDb)
	if err != nil {
		return nil, err
	}

	return tokenDb, nil
}

func (db *MongoDB) TokenByProfileId(id ID) (*Token, error) {
	tokenDb := &Token{}

	collection := db.collections[TokensDB]

	res := collection.FindOne(db.ctx, bson.D{
		{"profile", id},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(tokenDb)
	if err != nil {
		return nil, err
	}

	return tokenDb, nil
}

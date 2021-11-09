package db

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"go.mongodb.org/mongo-driver/bson"
)

type Token struct {
	Id        ID     `bson:"_id"`
	ProfileId ID     `bson:"profile"`
	Token     string `bson:"token"`
}

func (t Token) Audience() []string {
	panic("implement me")
}

func (t Token) Expiration() time.Time {
	panic("implement me")
}

func (t Token) IssuedAt() time.Time {
	panic("implement me")
}

func (t Token) Issuer() string {
	panic("implement me")
}

func (t Token) JwtID() string {
	panic("implement me")
}

func (t Token) NotBefore() time.Time {
	panic("implement me")
}

func (t Token) Subject() string {
	panic("implement me")
}

func (t Token) PrivateClaims() map[string]interface{} {
	panic("implement me")
}

func (t Token) Get(s string) (interface{}, bool) {
	panic("implement me")
}

func (t Token) Set(s string, i interface{}) error {
	panic("implement me")
}

func (t Token) Iterate(ctx context.Context) jwt.Iterator {
	panic("implement me")
}

func (t Token) Walk(ctx context.Context, visitor jwt.Visitor) error {
	panic("implement me")
}

func (t Token) AsMap(ctx context.Context) (map[string]interface{}, error) {
	panic("implement me")
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

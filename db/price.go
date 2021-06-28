package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Price struct {
	Timestamp int64   `bson:"timestamp"`
	Currency  string  `bson:"currency"`
	Price     float32 `bson:"price"`
}

func (db *MongoDB) Prices(currency string, startTime int64, endTime int64) ([]Price, error) {
	collection := db.collections[PricesDB]

	var price []Price
	res, err := collection.Find(db.ctx, bson.D{
		{"currency", currency},
		{"timestamp", bson.M{"$gte": startTime}},
		{"timestamp", bson.M{"$lte": endTime}},
	})
	if err != nil {
		return nil, err
	}

	err = res.Decode(price)
	if err != nil {
		return nil, err
	}

	return price, nil
}

func (db *MongoDB) LastPriceByCurrency(currency string) (*Price, error) {
	opt := options.FindOne()
	opt.SetSort(bson.D{{"timestamp", -1}})

	collection := db.collections[PricesDB]

	var price *Price
	res := collection.FindOne(db.ctx, bson.D{
		{"currency", currency},
	}, opt)
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(price)
	if err != nil {
		return nil, err
	}

	return price, nil
}

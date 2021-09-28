package db

import (
	"fractapp-server/types"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

type Profile struct {
	Id          ID                        `bson:"_id"`
	AuthId      string                    `bson:"auth_id"`
	Name        string                    `bson:"name"`
	Username    string                    `bson:"username"`
	PhoneNumber string                    `bson:"phone_number"`
	Email       string                    `bson:"email"`
	AvatarExt   string                    `bson:"avatar_ext"`
	LastUpdate  int64                     `bson:"last_update"`
	IsChatBot   bool                      `bson:"is_chat_bot"`
	Addresses   map[types.Network]Address `bson:"addresses"`
}

type Address struct {
	Address string `bson:"address"`
}

func (db *MongoDB) profileBy(property string, value interface{}) (*Profile, error) {
	p := &Profile{}

	collection := db.collections[ProfilesDB]
	res := collection.FindOne(db.ctx, bson.D{
		{property, value},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(p)
	if err != nil {
		return nil, err
	}

	return p, err
}

func (db *MongoDB) SearchUsersByUsername(value string, limit int64) ([]Profile, error) {
	profiles := make([]Profile, 0)

	opt := options.Find()
	opt.SetLimit(limit)

	collection := db.collections[ProfilesDB]
	res, err := collection.Find(db.ctx, bson.D{
		{"username", primitive.Regex{Pattern: "^" + value}},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &profiles)
	if err != nil {
		return nil, err
	}

	return profiles, err
}
func (db *MongoDB) SearchUsersByEmail(email string) (*Profile, error) {
	p, err := db.profileBy("email", email)
	if err != nil {
		return nil, err
	}

	return p, err
}

func (db *MongoDB) ProfileById(id ID) (*Profile, error) {
	p, err := db.profileBy("_id", id)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) ProfileByAuthId(authId string) (*Profile, error) {
	p, err := db.profileBy("auth_id", authId)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) ProfileByUsername(username string) (*Profile, error) {
	p, err := db.profileBy("username", username)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) ProfileByAddress(network types.Network, address string) (*Profile, error) {
	p, err := db.profileBy("addresses."+strconv.FormatInt(int64(network), 10), address)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) ProfileByPhoneNumber(phoneNumber string) (*Profile, error) {
	p, err := db.profileBy("phone_number", phoneNumber)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) ProfileByEmail(email string) (*Profile, error) {
	p, err := db.profileBy("email", email)
	if err != nil {
		return nil, err
	}

	return p, err
}
func (db *MongoDB) IsUsernameExist(username string) (bool, error) {
	_, err := db.profileBy("username", username)
	if err != ErrNoRows && err != nil {
		return false, err
	}

	if err != ErrNoRows {
		return false, nil
	}

	return true, nil
}

func (db *MongoDB) ProfilesCount() (int64, error) {
	collection := db.collections[ProfilesDB]
	count, err := collection.CountDocuments(db.ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return count, nil
}

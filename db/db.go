package db

import (
	"context"
	"errors"
	"fractapp-server/notification"
	"fractapp-server/types"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRef struct {
	Ref interface{} `bson:"$ref"`
	ID  ID          `bson:"$id"`
	DB  interface{} `bson:"$db"`
}

type ID primitive.ObjectID

var (
	ErrNoRows            = mongo.ErrNilDocument //TODO
	InvalidCollectionErr = errors.New("invalid collection name")

	AuthDB        name = "auth"
	ContactsDB    name = "contacts"
	MessagesDB    name = "messages"
	PricesDB      name = "prices"
	ProfilesDB    name = "profiles"
	SubscribersDB name = "subscribers"
	TokensDB      name = "tokens"
)

type name string

type DB interface {
	AuthByValue(value string, codeType notification.NotificatorType) (*Auth, error)

	AllContacts(profileId ID) ([]Contact, error)
	AllMatchContacts(id ID) ([]Profile, error)

	MessagesByReceiver(receiver ID) ([]Message, error)
	MessagesBySenderAndReceiver(sender ID, receiver ID) ([]Message, error)
	UpdateDeliveredMessage(id primitive.ObjectID) error

	Prices(currency string, startTime int64, endTime int64) ([]Price, error)
	LastPriceByCurrency(currency string) (*Price, error)

	SearchUsersByUsername(value string, limit int64) ([]Profile, error)
	SearchUsersByEmail(email string) (*Profile, error)

	ProfileById(id ID) (*Profile, error)
	ProfileByAuthId(authId string) (*Profile, error)
	ProfileByUsername(username string) (*Profile, error)
	ProfileByAddress(network types.Network, address string) (*Profile, error)
	ProfileByPhoneNumber(phoneNumber string) (*Profile, error)
	ProfileByEmail(email string) (*Profile, error)
	IsUsernameExist(username string) (bool, error)
	ProfilesCount() (int64, error)

	SubscribersCountByToken(token string) (int64, error)
	SubscriberByProfileId(id ID) (*Subscriber, error)

	TokenByValue(token string) (*Token, error)
	TokenByProfileId(id ID) (*Token, error)

	Insert(value interface{}) error
	InsertMany(values []interface{}) error
	UpdateByPK(id ID, value interface{}) error
}

type MongoDB struct {
	ctx         context.Context
	client      *mongo.Client
	database    *mongo.Database
	collections map[name]*mongo.Collection
}

func NewMongoDB(ctx context.Context, client *mongo.Client) (*MongoDB, error) {
	database := client.Database("fractapp")

	collection := database.Collection(string(AuthDB), nil)
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "value"}},
			Options: options.Index().SetUnique(true),
		},
	)

	collection = database.Collection(string(ProfilesDB), nil)
	_, err = collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "auth_id"}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, err
	}

	collections := map[name]*mongo.Collection{
		AuthDB:        database.Collection(string(AuthDB)),
		ContactsDB:    database.Collection(string(ContactsDB)),
		MessagesDB:    database.Collection(string(MessagesDB)),
		PricesDB:      database.Collection(string(PricesDB)),
		ProfilesDB:    database.Collection(string(ProfilesDB)),
		SubscribersDB: database.Collection(string(SubscribersDB)),
		TokensDB:      database.Collection(string(TokensDB)),
	}

	return &MongoDB{
		ctx:         ctx,
		client:      client,
		database:    database,
		collections: collections,
	}, nil
}

func (db *MongoDB) collection(value interface{}) (*mongo.Collection, error) {
	switch value.(type) {
	case Auth:
		return db.collections[AuthDB], nil
	case Contact:
		return db.collections[ContactsDB], nil
	case Message:
		return db.collections[MessagesDB], nil
	case Price:
		return db.collections[PricesDB], nil
	case Profile:
		return db.collections[ProfilesDB], nil
	case Subscriber:
		return db.collections[SubscribersDB], nil
	case Token:
		return db.collections[TokensDB], nil
	default:
		return nil, InvalidCollectionErr
	}
}

func (db *MongoDB) Insert(value interface{}) error {
	collection, err := db.collection(value)
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(db.ctx, value)
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoDB) InsertMany(values []interface{}) error {
	collection, err := db.collection(values[0])
	if err != nil {
		return err
	}

	_, err = collection.InsertMany(db.ctx, values)
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoDB) UpdateByPK(id ID, value interface{}) error {
	collection, err := db.collection(value)
	if err != nil {
		return err
	}

	_, err = collection.UpdateByID(db.ctx, id, value)
	if err != nil {
		return err
	}

	return nil
}

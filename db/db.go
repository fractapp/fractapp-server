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
	ErrNoRows            = mongo.ErrNoDocuments
	InvalidCollectionErr = errors.New("invalid collection name")

	AuthDB          name = "auth"
	ContactsDB      name = "contacts"
	MessagesDB      name = "messages"
	PricesDB        name = "prices"
	ProfilesDB      name = "profiles"
	SubscribersDB   name = "subscribers"
	TokensDB        name = "tokens"
	TransactionsDB  name = "transactions"
	NotificationsDB name = "notifications"
)

type name string

type DB interface {
	AuthByValue(value string, codeType notification.NotificatorType) (*Auth, error)

	AllContacts(profileId ID) ([]Contact, error)
	AllMatchContacts(id ID) ([]Profile, error)

	MessageById(id ID) (*Message, error)
	MessagesByReceiver(receiver ID) ([]Message, error)
	MessagesBySenderAndReceiver(sender ID, receiver ID) ([]Message, error)

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

	TransactionById(id ID) (*Transaction, error)
	TransactionByTxIdAndOwner(txId string, owner ID) (*Transaction, error)
	TransactionsByOwner(ownerAddress string, currency types.Currency) ([]Transaction, error)

	NotificationsByUserId(userId ID) ([]Notification, error)
	UndeliveredNotificationsByUserId(userId ID) ([]Notification, error)
	UndeliveredNotifications(maxTimestamp int64) ([]Notification, error)
	NotificationsByUserIdAndType(userId ID, nType NotificationType) ([]Notification, error)

	Insert(value interface{}) error
	InsertMany(values []interface{}) error
	UpdateByPK(Id ID, value interface{}) error
}

type MongoDB struct {
	ctx         context.Context
	client      *mongo.Client
	database    *mongo.Database
	collections map[name]*mongo.Collection
}

func NewId() ID {
	return ID(primitive.NewObjectID())
}
func NewMongoDB(ctx context.Context, client *mongo.Client) (*MongoDB, error) {
	database := client.Database("fractapp")

	collection := database.Collection(string(AuthDB), nil)
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "value", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	collection = database.Collection(string(ProfilesDB), nil)
	_, err = collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "auth_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, err
	}

	collections := map[name]*mongo.Collection{
		AuthDB:          database.Collection(string(AuthDB)),
		ContactsDB:      database.Collection(string(ContactsDB)),
		MessagesDB:      database.Collection(string(MessagesDB)),
		PricesDB:        database.Collection(string(PricesDB)),
		ProfilesDB:      database.Collection(string(ProfilesDB)),
		SubscribersDB:   database.Collection(string(SubscribersDB)),
		TokensDB:        database.Collection(string(TokensDB)),
		TransactionsDB:  database.Collection(string(TransactionsDB)),
		NotificationsDB: database.Collection(string(NotificationsDB)),
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
	case *Auth:
		return db.collections[AuthDB], nil

	case Contact:
		return db.collections[ContactsDB], nil
	case *Contact:
		return db.collections[ContactsDB], nil

	case Message:
		return db.collections[MessagesDB], nil
	case *Message:
		return db.collections[MessagesDB], nil

	case Price:
		return db.collections[PricesDB], nil
	case *Price:
		return db.collections[PricesDB], nil

	case Profile:
		return db.collections[ProfilesDB], nil
	case *Profile:
		return db.collections[ProfilesDB], nil

	case Subscriber:
		return db.collections[SubscribersDB], nil
	case *Subscriber:
		return db.collections[SubscribersDB], nil

	case Token:
		return db.collections[TokensDB], nil
	case *Token:
		return db.collections[TokensDB], nil

	case Transaction:
		return db.collections[TransactionsDB], nil
	case *Transaction:
		return db.collections[TransactionsDB], nil

	case Notification:
		return db.collections[NotificationsDB], nil
	case *Notification:
		return db.collections[NotificationsDB], nil
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

func (db *MongoDB) UpdateByPK(Id ID, value interface{}) error {
	collection, err := db.collection(value)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(db.ctx, bson.D{{"_id", Id}}, bson.D{{"$set", value}})
	if err != nil {
		return err
	}

	return nil
}

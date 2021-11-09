package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationType int32

const (
	TransactionNotificationType NotificationType = iota
	MessageNotificationType
)

type Notification struct {
	Id               ID               `bson:"_id"`
	Type             NotificationType `bson:"type"`
	Title            string           `bson:"title"`
	Message          string           `bson:"message"`
	TargetId         ID               `bson:"target_id"`
	UserId           ID               `bson:"user_id"`
	FirebaseNotified bool             `bson:"firebase_notified"`
	Delivered        bool             `bson:"delivered"`
	Timestamp        int64            `bson:"timestamp"`
}

func (db *MongoDB) NotificationsByUserId(userId ID) ([]Notification, error) {
	collection := db.collections[NotificationsDB]
	notifications := make([]Notification, 0)

	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	res, err := collection.Find(db.ctx, bson.D{
		{"user_id", userId},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, err
}

func (db *MongoDB) UndeliveredNotifications(maxTimestamp int64) ([]Notification, error) {
	collection := db.collections[NotificationsDB]

	notifications := make([]Notification, 0)
	res, err := collection.Find(db.ctx, bson.D{
		{"timestamp", bson.M{"$lte": maxTimestamp}},
		{"delivered", false},
		{"firebase_notified", false},
	})
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (db *MongoDB) UndeliveredNotificationsByUserId(userId ID) ([]Notification, error) {
	collection := db.collections[NotificationsDB]
	notifications := make([]Notification, 0)

	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	res, err := collection.Find(db.ctx, bson.D{
		{"user_id", userId},
		{"delivered", false},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, err
}

func (db *MongoDB) NotificationsByUserIdAndType(userId ID, nType NotificationType) ([]Notification, error) {
	collection := db.collections[NotificationsDB]
	notifications := make([]Notification, 0)

	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	res, err := collection.Find(db.ctx, bson.D{
		{"user_id", userId},
		{"type", nType},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, err
}

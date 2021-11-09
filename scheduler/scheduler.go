package scheduler

import (
	"context"
	"fractapp-server/db"
	"fractapp-server/push"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"
)

func Start(database db.DB, notificator push.Notificator, ctx context.Context) {
	for {
		select {
		case <-time.After(2 * time.Second):
			err := call(database, notificator)
			if err != nil {
				log.Errorf("error: %s \n", err.Error())
			}
		case <-ctx.Done():
			log.Println("scheduler shutdown")
		}
		time.Sleep(time.Minute)
	}
}

func call(database db.DB, notificator push.Notificator) error {
	now := time.Now()
	maxTimestamp := now.Add(-1 * time.Minute)

	notifications, err := database.UndeliveredNotifications(maxTimestamp.Unix())
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		log.Infof("scheduler - notification: %s \n", primitive.ObjectID(notification.Id).Hex())

		firebaseToken, err := database.SubscriberByProfileId(notification.UserId)
		if err != nil && err != db.ErrNoRows {
			return err
		} else if err == db.ErrNoRows {
			log.Infof("err token no found (userId): %s \n", primitive.ObjectID(notification.UserId).Hex())

			notification.FirebaseNotified = true
			err := database.UpdateByPK(notification.Id, &notification)
			if err != nil {
				return err
			}
			continue
		}

		err = notificator.Notify(notification.Title, notification.Message, firebaseToken.Token)
		if err != nil {
			return err
		}
		notification.FirebaseNotified = true
		err = database.UpdateByPK(notification.Id, &notification)
		if err != nil {
			return err
		}
	}

	return nil
}

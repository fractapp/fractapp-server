package scheduler

import (
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	notificatorMock "fractapp-server/mocks/push"
	"testing"
	"time"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestCall(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	mockNotificator := notificatorMock.NewMockNotificator(ctrl)

	timestampNow := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	unixTimestamp := timestampNow.Add(-1 * time.Minute).Unix()

	patchTimestamp := monkey.Patch(time.Now, func() time.Time { return timestampNow })
	defer patchTimestamp.Unpatch()

	notifications := []db.Notification{
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title",
			Message:          "message",
			TargetId:         db.NewId(),
			UserId:           db.NewId(),
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        10000,
		},
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         db.NewId(),
			UserId:           db.NewId(),
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        10000,
		},
	}
	mockDb.EXPECT().UndeliveredNotifications(unixTimestamp).Return(notifications, nil)

	subscriber := &db.Subscriber{
		Id:        db.NewId(),
		ProfileId: notifications[0].UserId,
		Token:     "token",
		Timestamp: 10000,
	}
	mockDb.EXPECT().SubscriberByProfileId(notifications[0].UserId).Return(subscriber, nil)
	notifications[0].FirebaseNotified = true
	mockDb.EXPECT().UpdateByPK(notifications[0].Id, &notifications[0])
	mockNotificator.EXPECT().Notify(notifications[0].Title, notifications[0].Message, subscriber.Token)

	mockDb.EXPECT().SubscriberByProfileId(notifications[1].UserId).Return(nil, db.ErrNoRows)
	notifications[1].FirebaseNotified = true
	mockDb.EXPECT().UpdateByPK(notifications[1].Id, &notifications[1])

	err := call(mockDb, mockNotificator)
	assert.Assert(t, err, nil)
}

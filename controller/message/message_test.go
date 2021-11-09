package message

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/types"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := NewController(dbMock.NewMockDB(ctrl))
	assert.Equal(t, c.MainRoute(), "/message")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case db.ErrNoRows:
		assert.Equal(t, w.Code, http.StatusNotFound)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}

func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)

	controller := NewController(dbMock.NewMockDB(ctrl))

	testErr(t, controller, db.ErrNoRows)
	testErr(t, controller, errors.New("any errors"))
}

var p = &db.Profile{
	Id:       db.NewId(),
	AuthId:   "authId",
	Name:     "nameSender",
	Username: "fractapper10",
	Addresses: map[types.Network]db.Address{
		types.Polkadot: {
			Address: "111111111111111111111111111111111HC1",
		},
		types.Kusama: {
			Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
		},
	},
}

func TestUnread(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	routeFn, err := controller.Handler("/unread")
	if err != nil {
		t.Fatal(err)
	}

	authId := "authId"
	mockDb.EXPECT().ProfileByAuthId(authId).Return(p, nil)
	senderP := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authIdSender",
		Username: "fractapper123",
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
			},
		},
	}
	messages := []*db.Message{
		{
			Id:         db.NewId(),
			Version:    2,
			Action:     "action",
			Value:      "value",
			Args:       nil,
			Rows:       nil,
			SenderId:   senderP.Id,
			ReceiverId: db.NewId(),
			Timestamp:  10001,
		},
		{
			Id:         db.NewId(),
			Version:    2,
			Action:     "action",
			Value:      "value",
			Args:       nil,
			Rows:       nil,
			SenderId:   senderP.Id,
			ReceiverId: db.NewId(),
			Timestamp:  10002,
		},
	}
	mockDb.EXPECT().UndeliveredNotificationsByUserId(p.Id).Return([]db.Notification{
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title",
			Message:          "message",
			TargetId:         messages[0].Id,
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        messages[0].Timestamp,
		},
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         messages[1].Id,
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        messages[1].Timestamp,
		},
		{
			Id:               db.NewId(),
			Type:             db.TransactionNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        messages[1].Timestamp,
		},
	}, nil)

	mockDb.EXPECT().MessageById(messages[0].Id).Return(messages[0], nil)
	mockDb.EXPECT().MessageById(messages[1].Id).Return(messages[1], nil)

	mockDb.EXPECT().ProfileById(senderP.Id).Return(senderP, nil)

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "auth_id", authId)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = routeFn(w, httpRq)
	assert.Assert(t, err == nil)

	messagesAndTxs := new(MessagesAndTxs)
	err = json.Unmarshal(w.Body.Bytes(), messagesAndTxs)
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, *messagesAndTxs, MessagesAndTxs{
		Messages: []MessageRs{
			{
				Id:        primitive.ObjectID(messages[0].Id).Hex(),
				Args:      messages[0].Args,
				Action:    Action(messages[0].Action),
				Version:   messages[0].Version,
				Value:     messages[0].Value,
				Rows:      messages[0].Rows,
				Sender:    senderP.AuthId,
				Receiver:  authId,
				Timestamp: messages[0].Timestamp,
			},
			{
				Id:        primitive.ObjectID(messages[1].Id).Hex(),
				Args:      messages[1].Args,
				Action:    Action(messages[1].Action),
				Version:   messages[1].Version,
				Value:     messages[1].Value,
				Rows:      messages[1].Rows,
				Sender:    senderP.AuthId,
				Receiver:  authId,
				Timestamp: messages[1].Timestamp,
			},
		},
		Users: map[string]profile.ShortUserProfile{
			"authIdSender": {
				Id:         senderP.AuthId,
				Name:       senderP.Name,
				Username:   senderP.Username,
				AvatarExt:  senderP.AvatarExt,
				LastUpdate: senderP.LastUpdate,
				IsChatBot:  senderP.IsChatBot,
				Addresses: map[types.Network]string{
					types.Polkadot: senderP.Addresses[types.Polkadot].Address,
					types.Kusama:   senderP.Addresses[types.Kusama].Address,
				},
			},
		},
	})
}

func TestRead(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	routeFn, err := controller.Handler("/read")
	if err != nil {
		t.Fatal(err)
	}

	notifications := []db.Notification{
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title",
			Message:          "message",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        1000,
		},
		{
			Id:               db.NewId(),
			Type:             db.MessageNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        1000,
		},
		{
			Id:               db.NewId(),
			Type:             db.TransactionNotificationType,
			Title:            "title2",
			Message:          "message2",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        1000,
		},
	}
	mockDb.EXPECT().UndeliveredNotificationsByUserId(p.Id).Return(notifications, nil)

	nOne := notifications[0]
	nOne.Delivered = true
	mockDb.EXPECT().UpdateByPK(nOne.Id, &nOne)

	nTwo := notifications[2]
	nTwo.Delivered = true
	mockDb.EXPECT().UpdateByPK(nTwo.Id, &nTwo)

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "profile_id", p.Id)

	rq := []string{
		primitive.ObjectID(notifications[0].TargetId).Hex(),
		primitive.ObjectID(notifications[2].TargetId).Hex(),
	}
	var body bytes.Buffer
	b, _ := json.Marshal(&rq)
	body.Write(b)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", &body)
	if err != nil {
		t.Fatal(err)
	}
	err = routeFn(w, httpRq)
	assert.Assert(t, err == nil)
}

func TestSend(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	routeFn, err := controller.Handler("/send")
	if err != nil {
		t.Fatal(err)
	}

	receiver := &db.Profile{
		Id:        db.NewId(),
		AuthId:    "authIdReceiver",
		Name:      "nameReceiver",
		Username:  "fractapper05",
		IsChatBot: true,
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
			},
		},
	}
	msg := MessageRq{
		Action:   "action",
		Receiver: receiver.AuthId,
		Args:     nil,
		Rows:     nil,
	}

	mockDb.EXPECT().ProfileByAuthId(p.AuthId).Return(p, nil)
	mockDb.EXPECT().ProfileByAuthId(receiver.AuthId).Return(receiver, nil)

	id := db.NewId()
	patchId := monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID { return primitive.ObjectID(id) })
	defer patchId.Unpatch()

	timestampNow := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	nanoTimestamp := timestampNow.UnixNano()
	unixTimestamp := timestampNow.Unix()

	patchTimestamp := monkey.Patch(time.Now, func() time.Time { return timestampNow })
	defer patchTimestamp.Unpatch()

	dbMessage := &db.Message{
		Id:         id,
		Value:      msg.Value,
		Action:     msg.Action,
		Version:    1,
		Args:       msg.Args,
		Rows:       msg.Rows,
		SenderId:   p.Id,
		ReceiverId: receiver.Id,
		Timestamp:  nanoTimestamp / int64(time.Millisecond),
	}
	mockDb.EXPECT().Insert(dbMessage).Return(nil)

	notification := &db.Notification{
		Id:               id,
		Type:             db.MessageNotificationType,
		Title:            p.Name,
		Message:          dbMessage.Value,
		TargetId:         dbMessage.Id,
		UserId:           dbMessage.ReceiverId,
		FirebaseNotified: false,
		Delivered:        false,
		Timestamp:        unixTimestamp,
	}
	mockDb.EXPECT().Insert(notification).Return(nil)

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "auth_id", p.AuthId)

	var body bytes.Buffer
	b, _ := json.Marshal(&msg)
	body.Write(b)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", &body)
	if err != nil {
		t.Fatal(err)
	}
	err = routeFn(w, httpRq)
	assert.Assert(t, err == nil)

	sendInfo := &SendInfo{}
	err = json.Unmarshal(w.Body.Bytes(), &sendInfo)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, sendInfo, &SendInfo{
		Timestamp: nanoTimestamp / int64(time.Millisecond),
	})
}

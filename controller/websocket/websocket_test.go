package websocket

import (
	"context"
	"errors"
	"fractapp-server/controller/info"
	"fractapp-server/controller/message"
	"fractapp-server/controller/middleware"
	internalMiddleware "fractapp-server/controller/middleware"
	"fractapp-server/controller/profile"
	"fractapp-server/controller/substrate"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/types"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"bou.ke/monkey"

	"github.com/go-chi/jwtauth"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

const txApiHost = "txApiHost"

func newController(t *testing.T) (*Controller, *dbMock.MockDB, *jwtauth.JWTAuth) {
	ctrl := gomock.NewController(t)
	mongoDB := dbMock.NewMockDB(ctrl)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	authMiddleware := internalMiddleware.New(mongoDB)
	c := NewController(mongoDB, tokenAuth, authMiddleware, txApiHost)

	return c, mongoDB, tokenAuth
}

func TestMainRoute(t *testing.T) {
	c, _, _ := newController(t)
	assert.Equal(t, c.MainRoute(), "/ws")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case middleware.InvalidAuthErr:
		assert.Equal(t, w.Code, http.StatusUnauthorized)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}

func TestReturnErr(t *testing.T) {
	controller, _, _ := newController(t)

	testErr(t, controller, middleware.InvalidAuthErr)
	testErr(t, controller, errors.New("any errors"))
}

func TestGetUsers(t *testing.T) {
	controller, mockDb, _ := newController(t)

	var p = &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId",
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

	users := []string{
		p.AuthId,
	}
	rq := &Rq{
		Method: getUsersMethod,
		Ids:    users,
	}
	mockDb.EXPECT().ProfileByAuthId(p.AuthId).Return(p, nil)

	rs := &WsResponse{
		Method: usersMethod,
		Value: map[string]*profile.ShortUserProfile{
			p.AuthId: {
				Id:         p.AuthId,
				Name:       p.Name,
				Username:   p.Username,
				AvatarExt:  p.AvatarExt,
				LastUpdate: p.LastUpdate,
				IsChatBot:  p.IsChatBot,
				Addresses: map[types.Network]string{
					types.Polkadot: p.Addresses[types.Polkadot].Address,
					types.Kusama:   p.Addresses[types.Kusama].Address,
				},
			},
		},
	}

	assert.DeepEqual(t, controller.getUsers(rq), rs)
}

func TestGetTxsStatuses(t *testing.T) {
	controller, _, _ := newController(t)

	txHash := "txHash"
	rq := &Rq{
		Method: getUsersMethod,
		Ids: []string{
			txHash,
		},
	}

	txHashMethod := ""
	txStatus := &profile.TxStatusRs{
		Hash:   "hash",
		Status: 0,
	}
	txStatusPatch := monkey.Patch(profile.TxStatus, func(txApiHost string, hash string) (*profile.TxStatusRs, error) {
		txHashMethod = txHash
		return txStatus, nil
	})
	defer txStatusPatch.Unpatch()

	rs := &WsResponse{
		Method: txsStatusesMethod,
		Value: []*profile.TxStatusRs{
			txStatus,
		},
	}

	assert.DeepEqual(t, controller.getTxsStatuses(rq, "authId"), rs)
	assert.DeepEqual(t, txHash, txHashMethod)
}

func TestSetDelivered(t *testing.T) {
	controller, mockDb, _ := newController(t)

	p := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId",
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
			Type:             db.TransactionNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        2000,
		},
	}
	rq := &Rq{
		Method: getUsersMethod,
		Ids: []string{
			primitive.ObjectID(notifications[0].Id).Hex(),
		},
	}

	mockDb.EXPECT().UndeliveredNotificationsByUserId(p.Id).Return(notifications, nil)
	newNotification := notifications[0]
	newNotification.Delivered = true
	mockDb.EXPECT().UpdateByPK(notifications[0].Id, &newNotification)

	controller.setDelivered(rq, p)
}

func TestNotifications(t *testing.T) {
	controller, mockDb, _ := newController(t)

	p := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId",
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
	pTwo := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authIdTwo",
		Username: "fractapper30",
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
			},
		},
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
			Type:             db.TransactionNotificationType,
			Title:            "title1",
			Message:          "message1",
			TargetId:         db.NewId(),
			UserId:           p.Id,
			FirebaseNotified: false,
			Delivered:        false,
			Timestamp:        2000,
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
			Timestamp:        2000,
		},
	}

	msg := &db.Message{
		Id:         notifications[0].TargetId,
		Version:    2,
		Action:     "action",
		Value:      "value",
		Args:       nil,
		Rows:       nil,
		SenderId:   pTwo.Id,
		ReceiverId: db.NewId(),
		Timestamp:  10001,
	}
	mockDb.EXPECT().UndeliveredNotificationsByUserId(p.Id).Return(notifications, nil)
	mockDb.EXPECT().MessageById(msg.Id).Return(msg, nil)
	mockDb.EXPECT().ProfileById(pTwo.Id).Return(pTwo, nil)

	mockDb.EXPECT().TransactionById(notifications[1].TargetId).Return(nil, db.ErrNoRows)
	newNotification := notifications[1]
	newNotification.Delivered = true
	mockDb.EXPECT().UpdateByPK(newNotification.Id, &newNotification).Return(nil)

	tx := &db.Transaction{
		Id:            notifications[2].TargetId,
		TxId:          "txId",
		Hash:          "txHash",
		Currency:      types.DOT,
		MemberAddress: "address",
		MemberId:      &pTwo.Id,
		Owner:         p.Id,
		Direction:     db.OutDirection,
		Status:        db.Success,
		Value:         "1000",
		Fee:           "100",
		Price:         125,
		Timestamp:     1001,
	}
	mockDb.EXPECT().TransactionById(tx.Id).Return(tx, nil)

	mockDb.EXPECT().LastPriceByCurrency(types.DOT.String()).Return(&db.Price{
		Timestamp: 10000,
		Currency:  "DOT",
		Price:     1001.1,
	}, nil).MaxTimes(1)
	mockDb.EXPECT().LastPriceByCurrency(types.KSM.String()).Return(&db.Price{
		Timestamp: 10005,
		Currency:  "KSM",
		Price:     1234.2358,
	}, nil).MaxTimes(1)

	var dataMock interface{}
	sendWsDataPatch := monkey.PatchInstanceMethod(reflect.TypeOf(controller), "SendWsData", func(c *Controller, data interface{}, id string) error {
		dataMock = data

		return nil
	})
	defer sendWsDataPatch.Unpatch()

	controller.notifications(p)

	assert.DeepEqual(t, dataMock, &WsResponse{
		Method: updateMethod,
		Value: &Update{
			Messages: []*message.MessageRs{
				{
					Id:        primitive.ObjectID(msg.Id).Hex(),
					Args:      msg.Args,
					Action:    message.Action(msg.Action),
					Version:   msg.Version,
					Value:     msg.Value,
					Rows:      msg.Rows,
					Sender:    pTwo.AuthId,
					Receiver:  p.AuthId,
					Timestamp: msg.Timestamp,
				},
			},
			Transactions: map[types.Currency][]*message.TransactionRs{
				types.DOT: {
					{
						Id:            tx.TxId,
						Hash:          tx.Hash,
						Currency:      tx.Currency,
						MemberAddress: tx.MemberAddress,
						Member:        &pTwo.AuthId,
						Direction:     tx.Direction,
						Action:        tx.Action,
						Status:        tx.Status,
						Value:         tx.Value,
						Fee:           tx.Value,
						Price:         tx.Price,
						Timestamp:     tx.Timestamp,
					},
				},
			},
			Users: map[string]profile.ShortUserProfile{
				pTwo.AuthId: {
					Id:         pTwo.AuthId,
					Name:       pTwo.Name,
					Username:   pTwo.Username,
					AvatarExt:  pTwo.AvatarExt,
					LastUpdate: pTwo.LastUpdate,
					IsChatBot:  pTwo.IsChatBot,
					Addresses: map[types.Network]string{
						types.Polkadot: pTwo.Addresses[types.Polkadot].Address,
						types.Kusama:   pTwo.Addresses[types.Kusama].Address,
					},
				},
			},
			Notifications: []string{primitive.ObjectID(notifications[0].Id).Hex(), primitive.ObjectID(notifications[2].Id).Hex()},
			Prices: []*info.Price{
				{
					Currency: types.DOT,
					Value:    1001.1,
				},
				{
					Currency: types.KSM,
					Value:    1234.2358,
				},
			},
		},
	})
}

func TestBalance(t *testing.T) {
	controller, _, _ := newController(t)

	p := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId",
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

	addresses := make([]string, 0)
	balancePatch := monkey.Patch(substrate.SubstrateBalance, func(txApiHost string, address string, currency types.Currency) (*substrate.Balance, error) {
		addresses = append(addresses, address)
		return &substrate.Balance{
			Total:         "1000",
			Transferable:  "2000",
			PayableForFee: "3000",
			Staking:       "4000",
		}, nil
	})
	defer balancePatch.Unpatch()

	var dataMock interface{}
	sendWsDataPatch := monkey.PatchInstanceMethod(reflect.TypeOf(controller), "SendWsData", func(c *Controller, data interface{}, id string) error {
		dataMock = data

		return nil
	})
	defer sendWsDataPatch.Unpatch()

	controller.balances(p)

	assert.Equal(t, len(addresses), 2)
	assert.Equal(t, addresses[0], p.Addresses[types.Polkadot].Address)
	assert.Equal(t, addresses[1], p.Addresses[types.Kusama].Address)
	assert.DeepEqual(t, dataMock, &WsResponse{
		Method: balancesMethod,
		Value: &Balances{
			Balances: map[types.Currency]*substrate.Balance{
				types.DOT: {
					Total:         "1000",
					Transferable:  "2000",
					PayableForFee: "3000",
					Staking:       "4000",
				},
				types.KSM: {
					Total:         "1000",
					Transferable:  "2000",
					PayableForFee: "3000",
					Staking:       "4000",
				},
			},
		},
	})
}

func TestJWTAuthPositive(t *testing.T) {
	controller, _, tokenAuth := newController(t)

	authId := "authId"
	tokenJWT, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": authId})
	if err != nil {
		t.Fatal(err)
	}

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, tokenJWT), "POST", "test?jwt="+tokenString, nil)
	if err != nil {
		t.Fatal(err)
	}

	profileId := db.NewId()

	tokenMock := ""
	pathchMiddleware := monkey.PatchInstanceMethod(reflect.TypeOf(controller.authMiddleware), "AuthWithJwt", func(m *middleware.AuthMiddleware, r *http.Request, findTokenFns func(r *http.Request) string) (string, db.ID, error) {
		tokenMock = findTokenFns(r)
		return authId, profileId, nil
	})
	defer pathchMiddleware.Unpatch()

	authIdOne, profileIdOne, err := controller.auth(rq)

	assert.Equal(t, tokenString, tokenMock)
	assert.Equal(t, authId, authIdOne)
	assert.Equal(t, profileId, profileIdOne)
}

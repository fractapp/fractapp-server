package subscriber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/push"
	"fractapp-server/types"
	"io/ioutil"
	"math/big"
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
	controller := NewController(dbMock.NewMockDB(ctrl))
	assert.Equal(t, controller.MainRoute(), "/")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(dbMock.NewMockDB(ctrl))

	testErr(t, controller, errors.New("any errors"))
}

func TestTransactionTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	routeFn, err := controller.Handler("/notify")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	address := "address"
	currency := types.DOT

	v := profile.Transaction{
		ID:        "id",
		Hash:      "hash",
		Action:    db.Transfer,
		Currency:  currency,
		To:        "to",
		From:      "from",
		Value:     "10000",
		Fee:       "1999123",
		Timestamp: 100023,
		Status:    db.Success,
	}
	rq := []profile.Transaction{
		v,
	}
	rqBytes, _ := json.Marshal(rq)
	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?address=%s&currency=%d", address, currency), ioutil.NopCloser(bytes.NewReader(rqBytes)))
	if err != nil {
		t.Fatal(err)
	}

	userFrom := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId2",
		Name:     "user2",
		Username: "fractapper2",
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "polkadot2",
			},
			types.Kusama: {
				Address: "kusama2",
			},
		},
	}
	userTo := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId1",
		Name:     "user1",
		Username: "fractapper1",
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "polkadot1",
			},
			types.Kusama: {
				Address: "kusama1",
			},
		},
	}

	id := db.NewId()
	patchId := monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID { return primitive.ObjectID(id) })
	defer patchId.Unpatch()

	price := float32(2.1)
	txTime := time.Unix(v.Timestamp/1000, 0)
	mockDb.EXPECT().Prices(
		currency.String(), txTime.
			Add(-15*time.Minute).Unix()*1000,
		txTime.
			Add(15*time.Minute).Unix()*1000,
	).
		Return([]db.Price{
			{
				Timestamp: v.Timestamp + (-12 * time.Minute).Milliseconds(),
				Currency:  currency.String(),
				Price:     12345,
			},
			{
				Timestamp: v.Timestamp + (1 * time.Minute).Milliseconds(),
				Currency:  currency.String(),
				Price:     price,
			},
		}, nil)

	mockDb.EXPECT().ProfileByAddress(v.Currency.Network(), v.From).Return(userFrom, nil)
	mockDb.EXPECT().ProfileByAddress(v.Currency.Network(), v.To).Return(userTo, nil)

	mockDb.EXPECT().TransactionByTxIdAndOwner(v.ID, userFrom.Id).Return(nil, db.ErrNoRows)
	mockDb.EXPECT().TransactionByTxIdAndOwner(v.ID, userTo.Id).Return(nil, db.ErrNoRows)

	senderTx := &db.Transaction{
		Id:            db.NewId(),
		TxId:          v.ID,
		Hash:          v.Hash,
		Currency:      v.Currency,
		MemberAddress: v.To,
		MemberId:      &userTo.Id,
		Owner:         userFrom.Id,
		Direction:     db.OutDirection,
		Status:        v.Status,
		Value:         v.Value,
		Fee:           v.Fee,
		Price:         price,
		Timestamp:     v.Timestamp,
	}

	receiverTx := &db.Transaction{
		Id:            db.NewId(),
		TxId:          v.ID,
		Hash:          v.Hash,
		Currency:      v.Currency,
		MemberAddress: v.From,
		MemberId:      &userFrom.Id,
		Owner:         userTo.Id,
		Direction:     db.InDirection,
		Status:        v.Status,
		Value:         v.Value,
		Fee:           v.Fee,
		Price:         price,
		Timestamp:     v.Timestamp,
	}

	mockDb.EXPECT().Insert(senderTx).Return(nil)
	mockDb.EXPECT().Insert(receiverTx).Return(nil)

	amount, _ := new(big.Int).SetString(v.Value, 10)
	fAmount, _ := currency.ConvertFromPlanck(amount).Float64()
	usdAmount := fAmount * float64(price)

	sentMsg := push.CreateMsg(push.Sent, fAmount, usdAmount, currency)

	receivedMsg := push.CreateMsg(push.Received, fAmount, usdAmount, currency)

	notifications := make([]interface{}, 0)
	notifications = append(notifications, &db.Notification{
		Id:        db.NewId(),
		Title:     userTo.Name,
		Message:   sentMsg,
		Type:      db.TransactionNotificationType,
		TargetId:  senderTx.Id,
		UserId:    userFrom.Id,
		Timestamp: time.Now().Unix(),
	})
	notifications = append(notifications, db.Notification{
		Id:        db.NewId(),
		Title:     userFrom.Name,
		Message:   receivedMsg,
		Type:      db.TransactionNotificationType,
		TargetId:  receiverTx.Id,
		UserId:    userTo.Id,
		Timestamp: time.Now().Unix(),
	})

	mockDb.EXPECT().InsertMany(notifications)

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
}

func TestTransactionStakingReward(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	routeFn, err := controller.Handler("/notify")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	address := "address"
	currency := types.DOT

	v := profile.Transaction{
		ID:        "id2",
		Hash:      "hash2",
		Action:    db.StakingReward,
		Currency:  currency,
		To:        "to",
		From:      "to",
		Value:     "300000",
		Fee:       "50000",
		Timestamp: 100023,
		Status:    db.Success,
	}
	rq := []profile.Transaction{
		v,
	}
	rqBytes, _ := json.Marshal(rq)
	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?address=%s&currency=%d", address, currency), ioutil.NopCloser(bytes.NewReader(rqBytes)))
	if err != nil {
		t.Fatal(err)
	}

	userTo := &db.Profile{
		Id:       db.NewId(),
		AuthId:   "authId1",
		Name:     "user1",
		Username: "fractapper1",
		Addresses: map[types.Network]db.Address{
			types.Polkadot: {
				Address: "polkadot1",
			},
			types.Kusama: {
				Address: "kusama1",
			},
		},
	}

	id := db.NewId()
	patchId := monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID { return primitive.ObjectID(id) })
	defer patchId.Unpatch()

	price := float32(2.1)
	txTime := time.Unix(v.Timestamp/1000, 0)
	mockDb.EXPECT().Prices(
		currency.String(), txTime.
			Add(-15*time.Minute).Unix()*1000,
		txTime.
			Add(15*time.Minute).Unix()*1000,
	).
		Return([]db.Price{
			{
				Timestamp: v.Timestamp + (-12 * time.Minute).Milliseconds(),
				Currency:  currency.String(),
				Price:     12345,
			},
			{
				Timestamp: v.Timestamp + (1 * time.Minute).Milliseconds(),
				Currency:  currency.String(),
				Price:     price,
			},
		}, nil)

	mockDb.EXPECT().ProfileByAddress(v.Currency.Network(), v.From).Return(userTo, nil)
	mockDb.EXPECT().ProfileByAddress(v.Currency.Network(), v.From).Return(userTo, nil)

	mockDb.EXPECT().TransactionByTxIdAndOwner(v.ID, userTo.Id).Return(nil, db.ErrNoRows)
	mockDb.EXPECT().TransactionByTxIdAndOwner(v.ID, userTo.Id).Return(nil, db.ErrNoRows)

	receiverTx := &db.Transaction{
		Id:            db.NewId(),
		TxId:          v.ID,
		Hash:          v.Hash,
		Currency:      v.Currency,
		MemberAddress: v.From,
		MemberId:      &userTo.Id,
		Owner:         userTo.Id,
		Direction:     db.InDirection,
		Status:        v.Status,
		Value:         v.Value,
		Fee:           v.Fee,
		Price:         price,
		Timestamp:     v.Timestamp,
	}

	mockDb.EXPECT().Insert(receiverTx).Return(nil)

	amount, _ := new(big.Int).SetString(v.Value, 10)
	fAmount, _ := currency.ConvertFromPlanck(amount).Float64()
	usdAmount := fAmount * float64(price)

	receivedMsg := push.CreateMsg(push.Received, fAmount, usdAmount, currency)

	mockDb.EXPECT().Insert(&db.Notification{
		Id:        db.NewId(),
		Title:     "Deposit payout",
		Message:   receivedMsg,
		Type:      db.TransactionNotificationType,
		TargetId:  receiverTx.Id,
		UserId:    userTo.Id,
		Timestamp: time.Now().Unix(),
	})

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
}

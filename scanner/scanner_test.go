package scanner

import (
	"errors"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"fractapp-server/mocks"
	"fractapp-server/types"
	"math/big"
	"testing"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestNewEventScanner(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	txNotificatorMock := mocks.NewMockTxNotificator(ctrl)
	adaptorMock := mocks.NewMockAdaptor(ctrl)
	e := NewBlockScanner(mockDb, "prefix", types.Polkadot, txNotificatorMock, adaptorMock)

	assert.Equal(t, e.db, mockDb)
	assert.Equal(t, e.prefix, "prefix")
	assert.Equal(t, e.network, types.Polkadot)
	assert.Equal(t, e.notificator, txNotificatorMock)
}

func TestScanNewHeightOne(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	txNotificatorMock := mocks.NewMockTxNotificator(ctrl)
	adaptorMock := mocks.NewMockAdaptor(ctrl)
	network := types.Polkadot
	e := NewBlockScanner(mockDb, "prefix", types.Polkadot, txNotificatorMock, adaptorMock)

	adaptorMock.EXPECT().WaitNewBlock().Return(uint64(101), nil)

	txs := []types.Tx{
		{
			Sender:     "sender1",
			Receiver:   "receiver1",
			FullAmount: big.NewInt(100000000000),
		},
		{
			Sender:     "sender2",
			Receiver:   "receiver2",
			FullAmount: big.NewInt(200000000000),
		},
	}
	subs := []db.Subscriber{
		{
			Address: "sender1",
			Token:   "token1",
			Network: network,
		},
		{
			Address: "sender2",
			Token:   "token2",
			Network: network,
		},
		{
			Address: "receiver1",
			Token:   "token3",
			Network: network,
		},
		{
			Address: "receiver2",
			Token:   "token4",
			Network: network,
		},
	}

	adaptorMock.EXPECT().Transfers(uint64(101)).Return(txs, nil)
	mockDb.EXPECT().SubscribersCount().Return(4, nil)
	mockDb.EXPECT().SubscribersByRange(0, 100000).Return(subs, nil)

	currency := network.Currency()
	mockDb.EXPECT().ProfileByAddress("sender1").Return(&db.Profile{
		Name: "senderName1",
	}, nil)
	mockDb.EXPECT().ProfileByAddress("receiver1").Return(&db.Profile{
		Name: "receiverName1",
	}, nil)
	mockDb.EXPECT().ProfileByAddress("sender2").Return(nil, db.ErrNoRows)
	mockDb.EXPECT().ProfileByAddress("receiver2").Return(nil, db.ErrNoRows)

	txNotificatorMock.EXPECT().Msg("receiverName1", firebase.Sent, float64(10), currency).Return("msg1")
	txNotificatorMock.EXPECT().Notify("msg1", "token1").Return(nil)

	txNotificatorMock.EXPECT().Msg("receiver2", firebase.Sent, float64(20), currency).Return("msg2")
	txNotificatorMock.EXPECT().Notify("msg2", "token2").Return(nil)

	txNotificatorMock.EXPECT().Msg("senderName1", firebase.Received, float64(10), currency).Return("msg3")
	txNotificatorMock.EXPECT().Notify("msg3", "token3").Return(nil)

	txNotificatorMock.EXPECT().Msg("sender2", firebase.Received, float64(20), currency).Return("msg4")
	txNotificatorMock.EXPECT().Notify("msg4", "token4").Return(nil)

	assert.Equal(t, e.scanNewHeight(100), uint64(101))
}

func TestScanNewHeightTwo(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	txNotificatorMock := mocks.NewMockTxNotificator(ctrl)
	adaptorMock := mocks.NewMockAdaptor(ctrl)
	e := NewBlockScanner(mockDb, "prefix", types.Polkadot, txNotificatorMock, adaptorMock)

	adaptorMock.EXPECT().WaitNewBlock().Return(uint64(100), nil)
	assert.Equal(t, e.scanNewHeight(101), uint64(101))
}

func TestScanNewHeightErr(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	txNotificatorMock := mocks.NewMockTxNotificator(ctrl)
	adaptorMock := mocks.NewMockAdaptor(ctrl)
	e := NewBlockScanner(mockDb, "prefix", types.Polkadot, txNotificatorMock, adaptorMock)

	adaptorMock.EXPECT().WaitNewBlock().Return(uint64(0), errors.New("error"))
	adaptorMock.EXPECT().Unsubscribe().Return()
	adaptorMock.EXPECT().Subscribe().Return(nil)
	assert.Equal(t, e.scanNewHeight(99), uint64(99))
}

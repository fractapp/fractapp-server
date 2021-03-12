package scanner

import (
	"fractapp-server/mocks"
	"fractapp-server/types"
	"testing"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestNewEventScanner(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	txNotificatorMock := mocks.NewMockTxNotificator(ctrl)

	e := NewEventScanner("host", mockDb, "prefix", types.Polkadot, txNotificatorMock)

	assert.Equal(t, e.host, "host")
	assert.Equal(t, e.db, mockDb)
	assert.Equal(t, e.prefix, "prefix")
	assert.Equal(t, e.network, types.Polkadot)
	assert.Equal(t, e.notificator, txNotificatorMock)
}

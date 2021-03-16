package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/mocks"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)

	controller := NewController(mocks.NewMockDB(ctrl))
	assert.Equal(t, controller.MainRoute(), "/notification")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case MaxAddressCountByTokenErr:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(mocks.NewMockDB(ctrl))

	testErr(t, controller, MaxAddressCountByTokenErr)
	testErr(t, controller, errors.New("any errors"))
}

func TestSubscribeForNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	timestamp := time.Date(2020, time.May, 19, 1, 2, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()
	rq := &UpdateTokenRq{
		PubKey:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Address:   "111111111111111111111111111111111HC1",
		Network:   types.Polkadot,
		Sign:      "sign",
		Token:     "token",
		Timestamp: timestamp.Unix(),
	}

	patchVerify := monkey.Patch(utils.Verify,
		func(pubKey [32]byte, msg string, hexSign string) error {
			return nil
		})
	defer patchVerify.Unpatch()

	mockDb.EXPECT().SubscribersCountByToken(rq.Token).Return(1, nil)
	mockDb.EXPECT().SubscriberByAddress(rq.Address).Return(nil, db.ErrNoRows)

	sub := &db.Subscriber{
		Address: rq.Address,
		Token:   rq.Token,
		Network: rq.Network,
	}
	mockDb.EXPECT().Insert(sub).Return(nil)

	subscribe, err := controller.Handler("/subscribe")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}

	err = subscribe(nil, httpRq)
	assert.Assert(t, err == nil)
}
func TestSubscribeForExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	timestamp := time.Date(2020, time.May, 19, 1, 2, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()
	rq := &UpdateTokenRq{
		PubKey:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Address:   "111111111111111111111111111111111HC1",
		Network:   types.Polkadot,
		Sign:      "sign",
		Token:     "token",
		Timestamp: timestamp.Unix(),
	}

	patchVerify := monkey.Patch(utils.Verify,
		func(pubKey [32]byte, msg string, hexSign string) error {
			return nil
		})
	defer patchVerify.Unpatch()

	mockDb.EXPECT().SubscribersCountByToken(rq.Token).Return(1, nil)

	sub := &db.Subscriber{
		Address: rq.Address,
		Token:   "token one",
		Network: rq.Network,
	}
	mockDb.EXPECT().SubscriberByAddress(rq.Address).Return(sub, nil)
	newSub := *sub
	newSub.Token = rq.Token
	mockDb.EXPECT().UpdateByPK(&newSub).Return(nil)

	subscribe, err := controller.Handler("/subscribe")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}

	err = subscribe(nil, httpRq)
	assert.Assert(t, err == nil)
}
func TestSubscribeSignTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	c := NewController(mockDb)

	rqTimestamp := time.Date(2020, time.May, 19, 1, 0, 0, 0, time.UTC)
	nowTimestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return nowTimestamp })
	defer patchTime.Unpatch()

	rq := &UpdateTokenRq{
		PubKey:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Address:   "111111111111111111111111111111111HC1",
		Network:   types.Polkadot,
		Sign:      "sign",
		Token:     "token",
		Timestamp: rqTimestamp.Unix(),
	}

	subscribe, err := c.Handler("/subscribe")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}

	err = subscribe(nil, httpRq)
	assert.Assert(t, err == controller.InvalidSignTimeErr)
}
func TestSubscribeInvalidAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	c := NewController(mockDb)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	rq := &UpdateTokenRq{
		PubKey:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Address:   "invalidAddress",
		Network:   types.Polkadot,
		Sign:      "sign",
		Token:     "token",
		Timestamp: timestamp.Unix(),
	}

	subscribe, err := c.Handler("/subscribe")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}

	err = subscribe(nil, httpRq)
	assert.Assert(t, err == controller.InvalidAddressErr)
}
func TestSubscribeMaxAddressCountByToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	c := NewController(mockDb)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	rq := &UpdateTokenRq{
		PubKey:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		Address:   "111111111111111111111111111111111HC1",
		Network:   types.Polkadot,
		Sign:      "sign",
		Token:     "token",
		Timestamp: timestamp.Unix(),
	}

	patchVerify := monkey.Patch(utils.Verify,
		func(pubKey [32]byte, msg string, hexSign string) error {
			return nil
		})
	defer patchVerify.Unpatch()

	mockDb.EXPECT().SubscribersCountByToken(rq.Token).Return(10, nil)

	subscribe, err := c.Handler("/subscribe")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}

	err = subscribe(nil, httpRq)
	assert.Assert(t, err == MaxAddressCountByTokenErr)
}

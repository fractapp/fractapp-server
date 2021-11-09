package substrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/types"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/docker/docker/pkg/ioutils"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

const txApiHost = "txApiHost"

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := NewController(dbMock.NewMockDB(ctrl), txApiHost)
	assert.Equal(t, c.MainRoute(), "/substrate")
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

	controller := NewController(dbMock.NewMockDB(ctrl), txApiHost)

	testErr(t, controller, db.ErrNoRows)
	testErr(t, controller, errors.New("any errors"))
}

func TestFee(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/fee")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	tx := "tx"
	network := types.Polkadot
	httpRq, err := http.NewRequest("POST", "http://127.0.0.1:80?tx="+tx+"&network="+fmt.Sprintf("%d", network), nil)
	if err != nil {
		t.Fatal(err)
	}

	feeInfo := FeeInfo{
		Fee: "10000",
	}
	rsByte, _ := json.Marshal(feeInfo)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, "txApiHost/substrate/fee?network=Polkadot&tx=tx")
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

func TestTransferFee(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/transfer/fee")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	network := types.Polkadot
	sender := "sender"
	receiver := "receiver"
	value := "value"
	isFullBalance := true

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?sender=%s&receiver=%s&value=%s&isFullBalance=%t&network=%d", sender, receiver, value, isFullBalance, network), nil)
	if err != nil {
		t.Fatal(err)
	}

	feeInfo := FeeInfo{
		Fee: "10000",
	}
	rsByte, _ := json.Marshal(feeInfo)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/substrate/transfer/fee?sender=%s&receiver=%s&value=%s&isFullBalance=%t&network=%s",
		txApiHost, sender, receiver, value, isFullBalance, network.String()))
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

func TestTxBase(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/txBase")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	network := types.Polkadot
	sender := "sender"

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?sender=%s&network=%d", sender, network), nil)
	if err != nil {
		t.Fatal(err)
	}

	rq := TxBase{
		BlockNumber: 1000,
		BlockHash:   "hash",
		Nonce:       123,
	}
	rsByte, _ := json.Marshal(rq)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/substrate/txBase/%s?network=%s",
		txApiHost, sender, network.String()))
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

func TestBase(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/base")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	network := types.Polkadot

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?network=%d", network), nil)
	if err != nil {
		t.Fatal(err)
	}

	rq := Base{
		GenesisHash:        "GenesisHash",
		Metadata:           "Metadata",
		SpecVersion:        12,
		TransactionVersion: 5,
	}
	rsByte, _ := json.Marshal(rq)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/substrate/base?network=%s",
		txApiHost, network.String()))
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

func TestBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/broadcast")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	tx := "tx"
	network := types.Polkadot

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?tx=%s&network=%d", tx, network), nil)
	if err != nil {
		t.Fatal(err)
	}

	rq := BroadcastResult{
		Hash: "hash",
	}
	rsByte, _ := json.Marshal(rq)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Post", func(client *http.Client, url string, contentType string, b io.Reader) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/substrate/broadcast?tx=%s&network=%s",
		txApiHost, tx, network.String()))
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

func TestBalance(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/balance")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	address := "address"
	currency := types.DOT

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?currency=%d&address=%s", currency, address), nil)
	if err != nil {
		t.Fatal(err)
	}

	rq := Balance{
		Total:         "1000",
		Transferable:  "2000",
		PayableForFee: "3000",
		Staking:       "4000",
	}
	rsByte, _ := json.Marshal(rq)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	mockUrl := ""
	httpPatch := monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		mockUrl = url
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil
	})
	defer httpPatch.Unpatch()

	err = routeFn(w, httpRq)
	assert.Assert(t, err, nil)
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/substrate/balance/%s?currency=%s",
		txApiHost, address, currency.String()))
	assert.DeepEqual(t, rsByte, w.Body.Bytes())
}

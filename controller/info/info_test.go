package info

import (
	"encoding/json"
	"errors"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := NewController(dbMock.NewMockDB(ctrl))
	assert.Equal(t, c.MainRoute(), "/info")
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

func TestInfo(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb)

	totalFn, err := controller.Handler("/total")
	if err != nil {
		t.Fatal(err)
	}

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

	w := httptest.NewRecorder()
	err = totalFn(w, nil)
	assert.Assert(t, err == nil)

	total := new(TotalInfo)
	err = json.Unmarshal(w.Body.Bytes(), total)
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, *total, TotalInfo{
		Prices: []Price{
			{
				Currency: types.DOT,
				Value:    1001.1,
			},
			{
				Currency: types.KSM,
				Value:    1234.2358,
			},
		},
	})
}

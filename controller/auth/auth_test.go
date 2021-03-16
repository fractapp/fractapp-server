package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/mocks"
	"fractapp-server/notification"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/go-chi/jwtauth"
	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	controller := NewController(mocks.NewMockDB(ctrl), nil, nil, tokenAuth)
	assert.Equal(t, controller.MainRoute(), "/auth")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case notification.InvalidEmailErr:
		fallthrough
	case InvalidCodeErr:
		fallthrough
	case notification.InvalidPhoneNumberErr:
		assert.Equal(t, w.Code, http.StatusNotFound)
	case InvalidSendTimeoutErr:
		assert.Equal(t, w.Code, http.StatusAccepted)
	case CodeExpiredErr:
		fallthrough
	case CodeUsedErr:
		fallthrough
	case InvalidNumberOfAttemptsErr:
		assert.Equal(t, w.Code, http.StatusTooManyRequests)
	case AddressExistErr:
		fallthrough
	case AccountExistErr:
		assert.Equal(t, w.Code, http.StatusForbidden)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	controller := NewController(mocks.NewMockDB(ctrl), mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	testErr(t, controller, notification.InvalidEmailErr)
	testErr(t, controller, InvalidCodeErr)
	testErr(t, controller, notification.InvalidPhoneNumberErr)
	testErr(t, controller, InvalidSendTimeoutErr)
	testErr(t, controller, InvalidNumberOfAttemptsErr)
	testErr(t, controller, CodeUsedErr)
	testErr(t, controller, CodeExpiredErr)
	testErr(t, controller, AddressExistErr)
	testErr(t, controller, AccountExistErr)
	testErr(t, controller, errors.New("any errors"))
}

func mockConfirmCode(mockDb *mocks.MockDB, value string, code string, notificatorType notification.NotificatorType) {
	expectAuthOne := &db.Auth{
		Value:     value,
		IsValid:   true,
		Code:      code,
		Attempts:  0,
		Count:     0,
		Timestamp: time.Now().Unix(),
		Type:      notificatorType,
		CheckType: notification.Auth,
	}

	mockDb.EXPECT().AuthByValue(value, notificatorType, notification.Auth).Return(expectAuthOne, nil)

	expectAuthTwo := *expectAuthOne
	expectAuthTwo.IsValid = false
	mockDb.EXPECT().UpdateByPK(&expectAuthTwo).Return(nil)
}

func TestConfirmPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb, mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	code := "123123"
	value := "phoneNumber"

	mockConfirmCode(mockDb, value, code, notification.SMS)
	err := controller.confirm(value, notification.SMS, notification.Auth, code)
	assert.Assert(t, err == nil)
}
func TestConfirmWithInvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb, mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	code := "123123"
	value := "phoneNumber"

	expectAuthOne := &db.Auth{
		Value:     value,
		IsValid:   true,
		Code:      "invalid",
		Attempts:  0,
		Count:     0,
		Timestamp: time.Now().Unix(),
		Type:      notification.SMS,
		CheckType: notification.Auth,
	}

	mockDb.EXPECT().AuthByValue(value, notification.SMS, notification.Auth).Return(expectAuthOne, nil)

	expectAuthTwo := *expectAuthOne
	expectAuthTwo.Attempts++
	mockDb.EXPECT().UpdateByPK(&expectAuthTwo).Return(nil)

	err := controller.confirm(value, notification.SMS, notification.Auth, code)
	assert.Assert(t, err == InvalidCodeErr)
}
func TestConfirmWithUsedCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb, mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	code := "123123"
	value := "phoneNumber"

	expectAuthOne := &db.Auth{
		Value:     value,
		IsValid:   false,
		Code:      code,
		Attempts:  0,
		Count:     0,
		Timestamp: time.Now().Unix(),
		Type:      notification.SMS,
		CheckType: notification.Auth,
	}

	mockDb.EXPECT().AuthByValue(value, notification.SMS, notification.Auth).Return(expectAuthOne, nil)

	expectAuthTwo := *expectAuthOne
	expectAuthTwo.IsValid = false
	mockDb.EXPECT().UpdateByPK(&expectAuthTwo).Return(nil)

	err := controller.confirm(value, notification.SMS, notification.Auth, code)
	assert.Assert(t, err == CodeUsedErr)
}
func TestConfirmWithMaxWrongCodeAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb, mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	code := "123123"
	value := "phoneNumber"

	expectAuthOne := &db.Auth{
		Value:     value,
		IsValid:   true,
		Code:      code,
		Attempts:  3,
		Count:     0,
		Timestamp: time.Now().Unix(),
		Type:      notification.SMS,
		CheckType: notification.Auth,
	}

	mockDb.EXPECT().AuthByValue(value, notification.SMS, notification.Auth).Return(expectAuthOne, nil)

	expectAuthTwo := *expectAuthOne
	expectAuthTwo.IsValid = false
	mockDb.EXPECT().UpdateByPK(&expectAuthTwo).Return(nil)

	err := controller.confirm(value, notification.SMS, notification.Auth, code)
	assert.Assert(t, err == InvalidNumberOfAttemptsErr)
}
func TestConfirmWithCodeExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb, mocks.NewMockNotificator(ctrl), mocks.NewMockNotificator(ctrl), tokenAuth)

	code := "123123"
	value := "phoneNumber"

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	authTimestamp := time.Date(2020, time.May, 19, 1, 0, 0, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	expectAuthOne := &db.Auth{
		Value:     value,
		IsValid:   true,
		Code:      code,
		Attempts:  0,
		Count:     0,
		Timestamp: authTimestamp.Unix(),
		Type:      notification.SMS,
		CheckType: notification.Auth,
	}

	mockDb.EXPECT().AuthByValue(value, notification.SMS, notification.Auth).Return(expectAuthOne, nil)

	expectAuthTwo := *expectAuthOne
	expectAuthTwo.IsValid = false
	mockDb.EXPECT().UpdateByPK(&expectAuthTwo).Return(nil)

	err := controller.confirm(value, notification.SMS, notification.Auth, code)
	assert.Assert(t, err == CodeExpiredErr)
}

func TestSendCodeForNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	sendCode, err := controller.Handler("/sendCode")
	if err != nil {
		t.Fatal(err)
	}

	rq := SendCodeRq{
		Type:      notification.Email,
		CheckType: notification.Auth,
		Value:     "test@test.com",
	}

	timestamp := time.Date(2020, time.May, 19, 1, 2, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockNotificator.EXPECT().Validate(rq.Value).Return(nil)

	code := "000000"
	patchCode := monkey.Patch(generateCode, func() string {
		return code
	})
	defer patchCode.Unpatch()

	mockDb.EXPECT().AuthByValue(rq.Value, rq.Type, rq.CheckType).Return(nil, db.ErrNoRows)

	auth := &db.Auth{
		Value:     rq.Value,
		Type:      rq.Type,
		IsValid:   true,
		CheckType: rq.CheckType,
		Code:      code,
		Timestamp: timestamp.Unix(),
		Count:     1,
		Attempts:  0,
	}

	mockDb.EXPECT().Insert(auth).Return(nil)
	mockNotificator.EXPECT().SendCode(rq.Value, code).Return(nil)

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	err = sendCode(nil, &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	})

	assert.Assert(t, err == nil)
}
func TestSendCodeForExistUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	sendCode, err := controller.Handler("/sendCode")
	if err != nil {
		t.Fatal(err)
	}

	rq := SendCodeRq{
		Type:      notification.Email,
		CheckType: notification.Auth,
		Value:     "test@test.com",
	}

	nowTimestamp := time.Date(2020, time.May, 19, 1, 10, 3, 4, time.UTC)
	rqTimestamp := time.Date(2020, time.May, 19, 1, 7, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return nowTimestamp })
	defer patchTime.Unpatch()

	newCode := "000000"
	patchCode := monkey.Patch(generateCode, func() string { return newCode })
	defer patchCode.Unpatch()

	existAuth := &db.Auth{
		Value:     rq.Value,
		Type:      rq.Type,
		IsValid:   false,
		CheckType: rq.CheckType,
		Code:      "old",
		Timestamp: rqTimestamp.Unix(),
		Count:     2,
		Attempts:  2,
	}

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockNotificator.EXPECT().Validate(rq.Value).Return(nil)
	mockDb.EXPECT().AuthByValue(rq.Value, rq.Type, rq.CheckType).Return(existAuth, nil)

	newAuth := *existAuth
	newAuth.Code = newCode
	newAuth.Timestamp = nowTimestamp.Unix()
	newAuth.Count = 3
	newAuth.Attempts = 0
	newAuth.IsValid = true

	mockDb.EXPECT().UpdateByPK(&newAuth).Return(nil)
	mockNotificator.EXPECT().SendCode(rq.Value, newCode).Return(nil)

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	err = sendCode(nil, &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	})

	assert.Assert(t, err == nil)
}
func TestSendCodeInvalidTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	sendCode, err := controller.Handler("/sendCode")
	if err != nil {
		t.Fatal(err)
	}

	rq := SendCodeRq{
		Type:      notification.Email,
		CheckType: notification.Auth,
		Value:     "test@test.com",
	}

	nowTimestamp := time.Date(2020, time.May, 19, 1, 10, 0, 0, time.UTC)
	rqTimestamp := time.Date(2020, time.May, 19, 1, 7, 1, 0, time.UTC) // < 3 minutes
	patchTime := monkey.Patch(time.Now, func() time.Time { return nowTimestamp })
	defer patchTime.Unpatch()

	newCode := "000000"
	patchCode := monkey.Patch(generateCode, func() string { return newCode })
	defer patchCode.Unpatch()

	existAuth := &db.Auth{
		Value:     rq.Value,
		Type:      rq.Type,
		IsValid:   false,
		CheckType: rq.CheckType,
		Code:      "old",
		Timestamp: rqTimestamp.Unix(),
		Count:     2,
		Attempts:  2,
	}

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockNotificator.EXPECT().Validate(rq.Value).Return(nil)
	mockDb.EXPECT().AuthByValue(rq.Value, rq.Type, rq.CheckType).Return(existAuth, nil)

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	err = sendCode(nil, &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	})

	assert.Assert(t, err == InvalidSendTimeoutErr)
}
func TestSendCodeMaxCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	sendCode, err := controller.Handler("/sendCode")
	if err != nil {
		t.Fatal(err)
	}

	rq := SendCodeRq{
		Type:      notification.Email,
		CheckType: notification.Auth,
		Value:     "test@test.com",
	}

	nowTimestamp := time.Date(2020, time.May, 19, 1, 10, 0, 0, time.UTC)
	rqTimestamp := time.Date(2020, time.May, 19, 1, 7, 0, 0, time.UTC) // < 3 minutes
	patchTime := monkey.Patch(time.Now, func() time.Time { return nowTimestamp })
	defer patchTime.Unpatch()

	newCode := "000000"
	patchCode := monkey.Patch(generateCode, func() string { return newCode })
	defer patchCode.Unpatch()

	existAuth := &db.Auth{
		Value:     rq.Value,
		Type:      rq.Type,
		IsValid:   false,
		CheckType: rq.CheckType,
		Code:      "old",
		Timestamp: rqTimestamp.Unix(),
		Count:     5,
		Attempts:  2,
	}

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockNotificator.EXPECT().Validate(rq.Value).Return(nil)
	mockDb.EXPECT().AuthByValue(rq.Value, rq.Type, rq.CheckType).Return(existAuth, nil)

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	err = sendCode(nil, &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	})

	assert.Assert(t, err == InvalidSendTimeoutErr)
}
func TestSendCodeResetCodeCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	sendCode, err := controller.Handler("/sendCode")
	if err != nil {
		t.Fatal(err)
	}

	rq := SendCodeRq{
		Type:      notification.Email,
		CheckType: notification.Auth,
		Value:     "test@test.com",
	}

	nowTimestamp := time.Date(2020, time.May, 19, 2, 1, 0, 0, time.UTC)
	rqTimestamp := time.Date(2020, time.May, 19, 1, 0, 0, 0, time.UTC) // < 3 minutes
	patchTime := monkey.Patch(time.Now, func() time.Time { return nowTimestamp })
	defer patchTime.Unpatch()

	newCode := "000000"
	patchCode := monkey.Patch(generateCode, func() string { return newCode })
	defer patchCode.Unpatch()

	existAuth := &db.Auth{
		Value:     rq.Value,
		Type:      rq.Type,
		IsValid:   false,
		CheckType: rq.CheckType,
		Code:      "old",
		Timestamp: rqTimestamp.Unix(),
		Count:     10,
		Attempts:  2,
	}

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockNotificator.EXPECT().Validate(rq.Value).Return(nil)
	mockDb.EXPECT().AuthByValue(rq.Value, rq.Type, rq.CheckType).Return(existAuth, nil)

	newAuth := *existAuth
	newAuth.Code = newCode
	newAuth.Timestamp = nowTimestamp.Unix()
	newAuth.Count = 1
	newAuth.Attempts = 0
	newAuth.IsValid = true

	mockDb.EXPECT().UpdateByPK(&newAuth).Return(nil)
	mockNotificator.EXPECT().SendCode(rq.Value, newCode).Return(nil)

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	err = sendCode(nil, &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	})

	assert.Assert(t, err == nil)
}

func TestSignForNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	code := "111111"
	rq := ConfirmAuthRq{
		Value: "phoneNumber",
		Type:  notification.SMS,
		Addresses: map[types.Network]Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signPolkadot",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signKusama",
			},
		},
		Code: code,
	}
	id := "userId"
	ctx := context.WithValue(context.Background(), "auth_id", id)

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockConfirmCode(mockDb, rq.Value, code, rq.Type)
	mockDb.EXPECT().ProfileById(id).Return(nil, db.ErrNoRows)

	timestamp := time.Date(2020, time.May, 19, 1, 2, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	patchVerify := monkey.Patch(utils.Verify,
		func(pubKey [32]byte, msg string, hexSign string) error {
			return nil
		})
	defer patchVerify.Unpatch()
	mockDb.EXPECT().AddressIsExist(rq.Addresses[types.Polkadot].Address).Return(false, nil).Times(1)
	mockDb.EXPECT().AddressIsExist(rq.Addresses[types.Kusama].Address).Return(false, nil).Times(1)

	profile := &db.Profile{
		Id:          id,
		IsMigratory: false,
		PhoneNumber: rq.Value,
	}
	addresses := []*db.Address{
		{
			Id:      id,
			Address: rq.Addresses[types.Polkadot].Address,
			Network: types.Polkadot,
		},
		{
			Id:      id,
			Address: rq.Addresses[types.Kusama].Address,
			Network: types.Kusama,
		},
	}
	mockDb.EXPECT().CreateProfile(ctx, gomock.Eq(profile), gomock.Eq(addresses)).Return(nil)
	mockDb.EXPECT().TokenById(id).Return("", db.ErrNoRows)

	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": id, "timestamp": timestamp.Unix()})
	if err != nil {
		t.Fatal(err)
	}
	mockDb.EXPECT().Insert(&db.Token{Token: tokenString, Id: id}).Return(nil)

	signIn, err := controller.Handler("/signIn")
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	httpRq.Header.Add("Sign-Timestamp", fmt.Sprintf("%d", timestamp.Unix()))
	httpRq.Header.Add("Auth-Key", "0x000000000000000000000000000000000000000000000000")

	w := httptest.NewRecorder()
	err = signIn(w, httpRq)

	assert.Assert(t, err == nil)

	token := &TokenRs{Token: tokenString}
	err = json.Unmarshal(w.Body.Bytes(), token)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, token.Token == tokenString)
}
func TestSignForInvalidSignTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	c := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	code := "111111"
	rq := ConfirmAuthRq{
		Value: "phoneNumber",
		Type:  notification.SMS,
		Addresses: map[types.Network]Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signPolkadot",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signKusama",
			},
		},
		Code: code,
	}
	id := "userId"
	ctx := context.WithValue(context.Background(), "auth_id", id)

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockConfirmCode(mockDb, rq.Value, code, rq.Type)
	mockDb.EXPECT().ProfileById(id).Return(nil, db.ErrNoRows)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	signIn, err := c.Handler("/signIn")
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	signTimestamp := time.Date(2020, time.May, 19, 1, 0, 0, 0, time.UTC)

	httpRq.Header.Add("Sign-Timestamp", fmt.Sprintf("%d", signTimestamp.Unix()))
	httpRq.Header.Add("Auth-Key", "0x000000000000000000000000000000000000000000000000")

	w := httptest.NewRecorder()
	err = signIn(w, httpRq)

	assert.Assert(t, err == controller.InvalidSignTimeErr)
}
func TestSignForExistUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	mockDb := mocks.NewMockDB(ctrl)
	mockNotificator := mocks.NewMockNotificator(ctrl)
	controller := NewController(mockDb, mockNotificator, mockNotificator, tokenAuth)

	code := "111111"
	rq := ConfirmAuthRq{
		Value: "email",
		Type:  notification.Email,
		Addresses: map[types.Network]Address{
			types.Polkadot: {
				Address: "111111111111111111111111111111111HC1",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signPolkadot",
			},
			types.Kusama: {
				Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
				PubKey:  "0x000000000000000000000000000000000000000000000000",
				Sign:    "signKusama",
			},
		},
		Code: code,
	}

	id := "userId"
	ctx := context.WithValue(context.Background(), "auth_id", id)

	mockNotificator.EXPECT().Format(rq.Value).Return(rq.Value)
	mockConfirmCode(mockDb, rq.Value, code, rq.Type)

	profile := &db.Profile{
		Id:          id,
		IsMigratory: false,
	}
	mockDb.EXPECT().ProfileById(id).Return(profile, nil)

	timestamp := time.Date(2020, time.May, 19, 1, 2, 3, 4, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	patchVerify := monkey.Patch(utils.Verify,
		func(pubKey [32]byte, msg string, hexSign string) error {
			return nil
		})
	defer patchVerify.Unpatch()
	mockDb.EXPECT().AddressIsExist(rq.Addresses[types.Polkadot].Address).Return(false, nil).Times(1)
	mockDb.EXPECT().AddressIsExist(rq.Addresses[types.Kusama].Address).Return(false, nil).Times(1)

	addresses := []db.Address{
		{
			Id:      id,
			Address: rq.Addresses[types.Polkadot].Address,
			Network: types.Polkadot,
		},
		{
			Id:      id,
			Address: rq.Addresses[types.Kusama].Address,
			Network: types.Kusama,
		},
	}
	mockDb.EXPECT().CreateProfile(ctx, gomock.Eq(profile), gomock.Eq(addresses)).Return(nil)

	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": id, "timestamp": timestamp.Unix()})
	if err != nil {
		t.Fatal(err)
	}
	mockDb.EXPECT().TokenById(id).Return(tokenString, nil)
	mockDb.EXPECT().AddressesById(id).Return(addresses, nil)

	newProfile := *profile
	newProfile.Email = rq.Value
	mockDb.EXPECT().UpdateByPK(&newProfile).Return(nil).Times(1)
	mockDb.EXPECT().Update(&db.Token{Token: tokenString, Id: id}, "id = ?", id).Return(nil).Times(1)

	signIn, err := controller.Handler("/signIn")
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	httpRq.Header.Add("Sign-Timestamp", fmt.Sprintf("%d", timestamp.Unix()))
	httpRq.Header.Add("Auth-Key", "0x000000000000000000000000000000000000000000000000")

	w := httptest.NewRecorder()
	err = signIn(w, httpRq)

	assert.Assert(t, err == nil)

	token := &TokenRs{Token: tokenString}
	err = json.Unmarshal(w.Body.Bytes(), token)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, token.Token == tokenString)
}

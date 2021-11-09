package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/db"
	dbMock "fractapp-server/mocks/db"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/ioutils"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(dbMock.NewMockDB(ctrl), "")
	assert.Equal(t, controller.MainRoute(), "/profile")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case db.ErrNoRows:
		assert.Equal(t, w.Code, http.StatusNotFound)
	case InvalidFileFormatErr:
		fallthrough
	case InvalidFileSizeErr:
		fallthrough
	case UsernameIsExistErr:
		fallthrough
	case InvalidPropertyErr:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	case UsernameNotFoundErr:
		assert.Equal(t, w.Code, http.StatusNotFound)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(dbMock.NewMockDB(ctrl), "")

	testErr(t, controller, db.ErrNoRows)
	testErr(t, controller, InvalidFileFormatErr)
	testErr(t, controller, InvalidFileSizeErr)
	testErr(t, controller, UsernameIsExistErr)
	testErr(t, controller, InvalidPropertyErr)
	testErr(t, controller, UsernameNotFoundErr)
	testErr(t, controller, errors.New("any errors"))
}

var addresses = map[types.Network]db.Address{
	types.Polkadot: {
		Address: "111111111111111111111111111111111HC1",
	},
	types.Kusama: {
		Address: "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp",
	},
}

var profile = &db.Profile{
	Id:        db.NewId(),
	AuthId:    "authId",
	Username:  "fractapper10",
	Addresses: addresses,
}

var user = &ShortUserProfile{
	Id:         profile.AuthId,
	Name:       profile.Name,
	Username:   profile.Username,
	AvatarExt:  profile.AvatarExt,
	LastUpdate: profile.LastUpdate,
	Addresses: map[types.Network]string{
		types.Polkadot: addresses[types.Polkadot].Address,
		types.Kusama:   addresses[types.Kusama].Address,
	},
}

func TestSearchByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	value := "value"
	mockDb.EXPECT().SearchUsersByEmail(value).Return(nil, db.ErrNoRows)
	mockDb.EXPECT().SearchUsersByUsername(value, int64(10)).Return([]db.Profile{*profile}, nil)

	search, err := controller.Handler("/search")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/search?value=" + value)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		URL: url,
	}
	returnErr := search(w, httpRq)

	var returnUsers []ShortUserProfile
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 1)
	assert.DeepEqual(t, returnUsers[0], *user)
}

func TestSearchByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	value := "value@test.com"

	mockDb.EXPECT().SearchUsersByEmail(value).Return(profile, nil)

	search, err := controller.Handler("/search")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/search?value=" + value + "&type=email")
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		URL: url,
	}
	returnErr := search(w, httpRq)

	var returnUsers []ShortUserProfile
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 1)
	assert.DeepEqual(t, returnUsers[0], *user)
}

func TestSearchMinSearchLength(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	value := "val"

	search, err := controller.Handler("/search")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/search?value=" + value)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		URL: url,
	}
	returnErr := search(w, httpRq)

	var returnUsers []ShortUserProfile
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 0)
}

func TestProfileInfoById(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	originalId := "      1231231231231231231231231231231231231231231231231231231231231234      "
	id := strings.Trim(originalId, " ")

	profileInfo, err := controller.Handler("/userInfo")
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().ProfileByAuthId(id).Return(profile, nil)

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/userInfo?id=" + id)
	if err != nil {
		t.Fatal(err)
	}
	httpRq := &http.Request{
		URL: url,
	}
	returnErr := profileInfo(w, httpRq)

	returnUser := &ShortUserProfile{}
	err = json.Unmarshal(w.Body.Bytes(), returnUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.DeepEqual(t, user, returnUser)
}

func TestMyProfile(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	id := "id"
	profileInfo, err := controller.Handler("/my")
	if err != nil {
		t.Fatal(err)
	}

	myProfile := &MyProfile{
		Id:          profile.AuthId,
		Name:        profile.Name,
		Username:    profile.Username,
		PhoneNumber: profile.PhoneNumber,
		Email:       profile.Email,
		IsMigratory: false,
		AvatarExt:   profile.AvatarExt,
		LastUpdate:  profile.LastUpdate,
	}

	mockDb.EXPECT().ProfileByAuthId(id).Return(profile, nil)

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "auth_id", id)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", nil)
	if err != nil {
		t.Fatal(err)
	}

	returnErr := profileInfo(w, httpRq)
	returnUser := &MyProfile{}
	err = json.Unmarshal(w.Body.Bytes(), returnUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.DeepEqual(t, myProfile, returnUser)
}

func TestUpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	id := "id"
	profileInfo, err := controller.Handler("/updateProfile")
	if err != nil {
		t.Fatal(err)
	}

	rq := &UpdateProfileRq{
		Name:     "New name",
		Username: "newusername",
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	profileArg := *profile
	mockDb.EXPECT().ProfileByAuthId(id).Return(&profileArg, nil)
	mockDb.EXPECT().IsUsernameExist(rq.Username).Return(false, nil)
	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	newProfile := *profile
	newProfile.Username = rq.Username
	newProfile.Name = rq.Name
	newProfile.LastUpdate = timestamp.Unix()

	mockDb.EXPECT().UpdateByPK(newProfile.Id, &newProfile).Return(nil)

	ctx := context.WithValue(context.Background(), "auth_id", id)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	err = profileInfo(nil, httpRq)

	assert.Assert(t, err == nil)
}

func TestMyContacts(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	profileInfo, err := controller.Handler("/contacts")
	if err != nil {
		t.Fatal(err)
	}

	contacts := []db.Contact{
		{
			Id:          db.NewId(),
			ProfileId:   profile.Id,
			PhoneNumber: "phoneOne",
		},
		{
			Id:          db.NewId(),
			ProfileId:   profile.Id,
			PhoneNumber: "phoneTwo",
		},
	}
	var stringContacts []string
	for _, v := range contacts {
		stringContacts = append(stringContacts, v.PhoneNumber)
	}

	mockDb.EXPECT().AllContacts(profile.Id).Return(contacts, nil)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	ctx := context.WithValue(context.Background(), "profile_id", profile.Id)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	err = profileInfo(w, httpRq)

	assert.Assert(t, err == nil)

	var returnContacts []string
	err = json.Unmarshal(w.Body.Bytes(), &returnContacts)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, returnContacts, stringContacts)
}

func TestMyMatchContacts(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	myMatchContacts, err := controller.Handler("/matchContacts")
	if err != nil {
		t.Fatal(err)
	}

	contacts := []db.Profile{
		*profile,
	}
	users := []ShortUserProfile{
		*user,
	}

	mockDb.EXPECT().ProfileById(profile.Id).Return(profile, nil)
	mockDb.EXPECT().AllMatchContacts(profile.Id).Return(contacts, nil)

	ctx := context.WithValue(context.Background(), "profile_id", profile.Id)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	err = myMatchContacts(w, httpRq)

	assert.Assert(t, err == nil)

	var returnUsers []ShortUserProfile
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, users, returnUsers)
}

func TestUploadMyContacts(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	existContacts := []db.Contact{
		{
			Id:          db.NewId(),
			ProfileId:   profile.Id,
			PhoneNumber: "contactOne",
		},
		{
			Id:          db.NewId(),
			ProfileId:   profile.Id,
			PhoneNumber: "contactTwo",
		},
		{
			Id:          db.NewId(),
			ProfileId:   profile.Id,
			PhoneNumber: "+12025550161",
		},
	}
	newContacts := []string{
		"+12025550198",
		"+12025550103",
		"+12025550161",
		"invalidContact",
	}

	uploadContacts, err := controller.Handler("/uploadContacts")
	if err != nil {
		t.Fatal(err)
	}

	id := db.NewId()
	patchId := monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID { return primitive.ObjectID(id) })
	defer patchId.Unpatch()

	mockDb.EXPECT().AllContacts(profile.Id).Return(existContacts, nil)

	contactOne := db.Contact{
		Id:          db.NewId(),
		ProfileId:   profile.Id,
		PhoneNumber: newContacts[0],
	}
	contactTwo := db.Contact{
		Id:          db.NewId(),
		ProfileId:   profile.Id,
		PhoneNumber: newContacts[1],
	}
	myContacts := []interface{}{
		contactOne,
		contactTwo,
	}

	mockDb.EXPECT().InsertMany(myContacts).Return(nil)

	ctx := context.WithValue(context.Background(), "profile_id", profile.Id)
	b, err := json.Marshal(newContacts)
	if err != nil {
		t.Fatal(err)
	}

	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	err = uploadContacts(nil, httpRq)

	assert.Assert(t, err == nil)
}

func TestFindUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	originalUsername := "UserName"
	username := strings.ToLower(originalUsername)
	findUsername, err := controller.Handler("/username")
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().IsUsernameExist(username).Return(true, nil)

	url, err := url.Parse("http://localhost:80/username?username=" + originalUsername)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		URL: url,
	}

	err = findUsername(nil, httpRq)

	assert.Assert(t, err == nil)
}

func TestFindUsernameNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	originalUsername := "UserName"
	username := strings.ToLower(originalUsername)
	findUsername, err := controller.Handler("/username")
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().IsUsernameExist(username).Return(false, nil)

	url, err := url.Parse("http://localhost:80/username?username=" + originalUsername)
	if err != nil {
		t.Fatal(err)
	}

	httpRq := &http.Request{
		URL: url,
	}

	err = findUsername(nil, httpRq)

	assert.Assert(t, err == UsernameNotFoundErr)
}

func TestUploadAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)

	id := "id"
	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	var body bytes.Buffer

	mwriter := multipart.NewWriter(&body)
	mwriter.WriteField("avatar", "MTIzYXNkZHNhMTIz")
	mwriter.WriteField("format", "image/jpeg")
	mwriter.Close()

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	patchWriteAvatar := monkey.Patch(utils.WriteAvatar, func(fileName string, decoded []byte) error { return nil })
	defer patchWriteAvatar.Unpatch()

	profileArg := *profile
	mockDb.EXPECT().ProfileByAuthId(id).Return(&profileArg, nil)

	newProfile := *profile
	newProfile.LastUpdate = timestamp.Unix()
	newProfile.AvatarExt = "jpeg"
	mockDb.EXPECT().UpdateByPK(newProfile.Id, &newProfile).Return(nil)

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), "auth_id", id), http.MethodPost, "http://localhost:80", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mwriter.FormDataContentType())

	uploadAvatar, err := controller.Handler("/uploadAvatar")
	if err != nil {
		t.Fatal(err)
	}
	err = uploadAvatar(nil, req)

	assert.Assert(t, err == nil)
}

func TestUploadAvatarInvalidFormat(t *testing.T) {
	ctrl := gomock.NewController(t)

	id := "id"
	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	var body bytes.Buffer

	mwriter := multipart.NewWriter(&body)
	mwriter.WriteField("avatar", "MTIzYXNkZHNhMTIz")
	mwriter.WriteField("format", "image/blabla")
	mwriter.Close()

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), "auth_id", id), http.MethodPost, "http://localhost:80", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mwriter.FormDataContentType())

	uploadAvatar, err := controller.Handler("/uploadAvatar")
	if err != nil {
		t.Fatal(err)
	}
	err = uploadAvatar(nil, req)

	assert.Assert(t, err == InvalidFileFormatErr)
}

func TestUpdateFirebaseTokenUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	rq := &UpdateFirebaseTokenRq{
		Token: "token",
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().SubscribersCountByToken(rq.Token).Return(int64(0), nil)
	sub := &db.Subscriber{
		Id:        db.NewId(),
		ProfileId: profile.Id,
		Token:     "token2",
		Timestamp: 1000,
	}
	mockDb.EXPECT().SubscriberByProfileId(profile.Id).Return(sub, nil)
	newSub := *sub
	newSub.Token = rq.Token
	mockDb.EXPECT().UpdateByPK(newSub.Id, &newSub)

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), "profile_id", profile.Id), http.MethodPost, "http://localhost:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	routeFn, err := controller.Handler("/firebase/update")
	if err != nil {
		t.Fatal(err)
	}
	err = routeFn(nil, req)

	assert.Assert(t, err == nil)
}

func TestUpdateFirebaseTokenCreate(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, "")

	rq := &UpdateFirebaseTokenRq{
		Token: "token",
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	id := db.NewId()
	patchId := monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID { return primitive.ObjectID(id) })
	defer patchId.Unpatch()

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	mockDb.EXPECT().SubscribersCountByToken(rq.Token).Return(int64(0), nil)
	mockDb.EXPECT().SubscriberByProfileId(profile.Id).Return(nil, db.ErrNoRows)
	sub := &db.Subscriber{
		Id:        db.NewId(),
		Token:     rq.Token,
		ProfileId: profile.Id,
		Timestamp: timestamp.Unix(),
	}
	mockDb.EXPECT().Insert(sub)

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), "profile_id", profile.Id), http.MethodPost, "http://localhost:80", ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	routeFn, err := controller.Handler("/firebase/update")
	if err != nil {
		t.Fatal(err)
	}
	err = routeFn(nil, req)

	assert.Assert(t, err == nil)
}

func TestTxStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	txApiHost := "txApiHost"
	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/transaction/status")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	hash := "hash"

	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?hash=%s", hash), nil)
	if err != nil {
		t.Fatal(err)
	}

	rs := TxStatusScannerApiRs{
		Status: 12,
	}
	rsByte, _ := json.Marshal(rs)
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
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/transaction/%s",
		txApiHost, hash))

	b, _ := json.Marshal(TxStatusRs{
		Hash:   hash,
		Status: rs.Status,
	})
	assert.DeepEqual(t, b, w.Body.Bytes())
}

func TestTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)

	txApiHost := "txApiHost"
	mockDb := dbMock.NewMockDB(ctrl)
	controller := NewController(mockDb, txApiHost)

	routeFn, err := controller.Handler("/transactions")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	address := "address"
	currency := types.DOT
	httpRq, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:80?address=%s&currency=%d", address, currency), nil)
	if err != nil {
		t.Fatal(err)
	}

	rs := []Transaction{
		{
			ID:        "id",
			Hash:      "hash",
			Action:    db.Transfer,
			Currency:  currency,
			To:        "to",
			From:      "from",
			Value:     "value",
			Fee:       "1999123",
			Timestamp: 100023,
			Status:    db.Success,
		},
	}
	rsByte, _ := json.Marshal(rs)
	r := ioutils.NewReadCloserWrapper(bytes.NewReader(rsByte), func() error {
		return nil
	})

	userTo := &db.Profile{
		Id:        db.NewId(),
		AuthId:    "authId12",
		Username:  "fractapper23",
		Addresses: addresses,
	}
	responseTxs := make([]OldTransactionRs, 0)
	for _, v := range rs {
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
					Price:     2,
				},
			}, nil)

		mockDb.EXPECT().ProfileByAddress(currency.Network(), v.From).Return(profile, nil)
		mockDb.EXPECT().ProfileByAddress(currency.Network(), v.To).Return(userTo, nil)

		responseTxs = append(responseTxs, OldTransactionRs{
			ID:   v.ID,
			Hash: v.Hash,

			Currency: v.Currency,

			From:     v.From,
			UserFrom: profile.AuthId,

			To:     v.To,
			UserTo: userTo.AuthId,

			Action: v.Action,
			Value:  v.Value,
			Price:  2,

			Fee:       v.Fee,
			Timestamp: v.Timestamp,
			Status:    v.Status,
		})
	}

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
	assert.DeepEqual(t, mockUrl, fmt.Sprintf("%s/transactions/%s?currency=%s",
		txApiHost, address, currency.String()))

	b, _ := json.Marshal(responseTxs)
	assert.DeepEqual(t, b, w.Body.Bytes())
}

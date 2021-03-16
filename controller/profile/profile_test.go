package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fractapp-server/db"
	"fractapp-server/mocks"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"

	"gotest.tools/assert"

	"github.com/golang/mock/gomock"
)

func TestMainRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(mocks.NewMockDB(ctrl))
	assert.Equal(t, controller.MainRoute(), "/profile")
}

func testErr(t *testing.T, controller *Controller, err error) {
	w := httptest.NewRecorder()
	controller.ReturnErr(err, w)

	switch err {
	case UsernameNotFoundErr:
		assert.Equal(t, w.Code, http.StatusNotFound)
	default:
		assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
func TestReturnErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	controller := NewController(mocks.NewMockDB(ctrl))

	testErr(t, controller, UsernameNotFoundErr)
	testErr(t, controller, errors.New("any errors"))
}

func TestSearchByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	value := "value"
	profiles := []db.Profile{
		{
			Id:          "idOne",
			Name:        "nameOne",
			Username:    "usernameOne",
			PhoneNumber: "phoneNumber",
			Email:       "email",
			IsMigratory: false,
			AvatarExt:   "png",
			LastUpdate:  123,
		},
	}
	addresses := []db.Address{
		{
			Id:      "addressId",
			Address: "address",
			Network: types.Polkadot,
		},
	}
	mockDb.EXPECT().SearchUsersByUsername(value, 10).Return(profiles, nil)
	mockDb.EXPECT().AddressesById(profiles[0].Id).Return(addresses, nil)

	user := UserProfileShort{
		Id:         profiles[0].Id,
		Name:       profiles[0].Name,
		Username:   profiles[0].Username,
		AvatarExt:  profiles[0].AvatarExt,
		LastUpdate: profiles[0].LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}
	for _, v := range addresses {
		user.Addresses[v.Network.Currency()] = v.Address
	}

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

	var returnUsers []UserProfileShort
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 1)
	assert.DeepEqual(t, returnUsers[0], user)
}
func TestSearchByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	value := "value@test.com"
	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	addresses := []db.Address{
		{
			Id:      "addressId",
			Address: "address",
			Network: types.Polkadot,
		},
	}

	mockDb.EXPECT().SearchUsersByEmail(value).Return(profile, nil)
	mockDb.EXPECT().AddressesById(profile.Id).Return(addresses, nil)

	user := UserProfileShort{
		Id:         profile.Id,
		Name:       profile.Name,
		Username:   profile.Username,
		AvatarExt:  profile.AvatarExt,
		LastUpdate: profile.LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}
	for _, v := range addresses {
		user.Addresses[v.Network.Currency()] = v.Address
	}

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

	var returnUsers []UserProfileShort
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 1)
	assert.DeepEqual(t, returnUsers[0], user)
}
func TestSearchMinSearchLength(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

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

	var returnUsers []UserProfileShort
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.Assert(t, len(returnUsers) == 0)
}

func TestProfileInfoById(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	originalId := "      123123      "
	id := strings.Trim(originalId, " ")

	profileInfo, err := controller.Handler("/info")
	if err != nil {
		t.Fatal(err)
	}

	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	user := &UserProfileShort{
		Id:         profile.Id,
		Name:       profile.Name,
		Username:   profile.Username,
		AvatarExt:  profile.AvatarExt,
		LastUpdate: profile.LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}
	addresses := []db.Address{
		{
			Id:      "addressId",
			Address: "address",
			Network: types.Polkadot,
		},
	}
	for _, v := range addresses {
		user.Addresses[v.Network.Currency()] = v.Address
	}

	mockDb.EXPECT().ProfileById(id).Return(profile, nil)
	mockDb.EXPECT().AddressesById(profile.Id).Return(addresses, nil)

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/info?id=" + id)
	if err != nil {
		t.Fatal(err)
	}
	httpRq := &http.Request{
		URL: url,
	}
	returnErr := profileInfo(w, httpRq)

	returnUser := &UserProfileShort{}
	err = json.Unmarshal(w.Body.Bytes(), returnUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.DeepEqual(t, user, returnUser)
}
func TestProfileInfoByAddress(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	originalAddress := "      address      "
	address := strings.Trim(originalAddress, " ")

	profileInfo, err := controller.Handler("/info")
	if err != nil {
		t.Fatal(err)
	}

	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	user := &UserProfileShort{
		Id:         profile.Id,
		Name:       profile.Name,
		Username:   profile.Username,
		AvatarExt:  profile.AvatarExt,
		LastUpdate: profile.LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}
	addresses := []db.Address{
		{
			Id:      "addressId",
			Address: "address",
			Network: types.Polkadot,
		},
	}
	for _, v := range addresses {
		user.Addresses[v.Network.Currency()] = v.Address
	}

	mockDb.EXPECT().ProfileByAddress(address).Return(profile, nil)
	mockDb.EXPECT().AddressesById(profile.Id).Return(addresses, nil)

	w := httptest.NewRecorder()
	url, err := url.Parse("http://localhost:80/info?address=" + address)
	if err != nil {
		t.Fatal(err)
	}
	httpRq := &http.Request{
		URL: url,
	}
	returnErr := profileInfo(w, httpRq)

	returnUser := &UserProfileShort{}
	err = json.Unmarshal(w.Body.Bytes(), returnUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, returnErr == nil)
	assert.DeepEqual(t, user, returnUser)
}
func TestMyProfile(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	id := "id"
	profileInfo, err := controller.Handler("/my")
	if err != nil {
		t.Fatal(err)
	}

	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	myProfile := &MyProfile{
		Id:          profile.Id,
		Name:        profile.Name,
		Username:    profile.Username,
		PhoneNumber: profile.PhoneNumber,
		Email:       profile.Email,
		IsMigratory: false,
		AvatarExt:   profile.AvatarExt,
		LastUpdate:  profile.LastUpdate,
	}

	mockDb.EXPECT().ProfileById(id).Return(profile, nil)

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

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	id := "id"
	profileInfo, err := controller.Handler("/updateProfile")
	if err != nil {
		t.Fatal(err)
	}

	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	rq := &UpdateProfileRq{
		Name:     "New Name",
		Username: "newusername",
	}
	b, err := json.Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().ProfileById(id).Return(profile, nil)
	mockDb.EXPECT().UsernameIsExist(rq.Username).Return(false, nil)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	newProfile := *profile
	newProfile.Username = rq.Username
	newProfile.Name = rq.Name
	newProfile.LastUpdate = timestamp.Unix()

	mockDb.EXPECT().UpdateByPK(&newProfile).Return(nil)

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

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	id := "id"
	profileInfo, err := controller.Handler("/contacts")
	if err != nil {
		t.Fatal(err)
	}

	contacts := []db.Contact{
		{
			Id:          id,
			PhoneNumber: "phoneOne",
		},
		{
			Id:          id,
			PhoneNumber: "phoneTwo",
		},
	}
	var stringContacts []string
	for _, v := range contacts {
		stringContacts = append(stringContacts, v.PhoneNumber)
	}

	mockDb.EXPECT().AllContacts(id).Return(contacts, nil)

	timestamp := time.Date(2020, time.May, 19, 1, 10, 1, 0, time.UTC)
	patchTime := monkey.Patch(time.Now, func() time.Time { return timestamp })
	defer patchTime.Unpatch()

	ctx := context.WithValue(context.Background(), "auth_id", id)
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

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	id := "id"
	myMatchContacts, err := controller.Handler("/matchContacts")
	if err != nil {
		t.Fatal(err)
	}

	profile := &db.Profile{
		Id:          id,
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	contacts := []db.Profile{
		{
			Id:          "idOne",
			Name:        "nameOne",
			Username:    "usernameOne",
			PhoneNumber: "phoneNumberOne",
			Email:       "emailOne",
			IsMigratory: false,
			AvatarExt:   "",
			LastUpdate:  123123,
		},
		{
			Id:          "idTwo",
			Name:        "nameTwo",
			Username:    "usernameTwo",
			PhoneNumber: "phoneNumberTwo",
			Email:       "emailTwo",
			IsMigratory: false,
			AvatarExt:   "png",
			LastUpdate:  123123,
		},
	}
	users := []UserProfileShort{
		{
			Id:         contacts[0].Id,
			Name:       contacts[0].Name,
			Username:   contacts[0].Username,
			AvatarExt:  contacts[0].AvatarExt,
			LastUpdate: contacts[0].LastUpdate,
			Addresses: map[types.Currency]string{
				types.DOT: "dotAddressOne",
				types.KSM: "ksmAddressOne",
			},
		},
		{
			Id:         contacts[1].Id,
			Name:       contacts[1].Name,
			Username:   contacts[1].Username,
			AvatarExt:  contacts[1].AvatarExt,
			LastUpdate: contacts[1].LastUpdate,
			Addresses: map[types.Currency]string{
				types.DOT: "dotAddressTwo",
				types.KSM: "ksmAddressTwo",
			},
		},
	}

	addressesOne := []db.Address{
		{
			Id:      users[0].Id,
			Address: users[0].Addresses[types.DOT],
			Network: types.Polkadot,
		},
		{
			Id:      users[0].Id,
			Address: users[0].Addresses[types.KSM],
			Network: types.Kusama,
		},
	}
	addressesTwo := []db.Address{
		{
			Id:      users[1].Id,
			Address: users[1].Addresses[types.DOT],
			Network: types.Polkadot,
		},
		{
			Id:      users[1].Id,
			Address: users[1].Addresses[types.KSM],
			Network: types.Kusama,
		},
	}

	mockDb.EXPECT().ProfileById(profile.Id).Return(profile, nil)
	mockDb.EXPECT().AllMatchContacts(profile.Id, profile.PhoneNumber).Return(contacts, nil)
	mockDb.EXPECT().AddressesById(users[0].Id).Return(addressesOne, nil).Times(1)
	mockDb.EXPECT().AddressesById(users[1].Id).Return(addressesTwo, nil).Times(1)

	ctx := context.WithValue(context.Background(), "auth_id", id)
	httpRq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:80", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	err = myMatchContacts(w, httpRq)

	assert.Assert(t, err == nil)

	var returnUsers []UserProfileShort
	err = json.Unmarshal(w.Body.Bytes(), &returnUsers)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, users, returnUsers)
}
func TestUploadMyContacts(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	id := "id"
	existContacts := []db.Contact{
		{
			Id:          id,
			PhoneNumber: "contactOne",
		},
		{
			Id:          id,
			PhoneNumber: "contactTwo",
		},
		{
			Id:          id,
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

	mockDb.EXPECT().AllContacts(id).Return(existContacts, nil)
	myContacts := []db.Contact{
		{
			Id:          id,
			PhoneNumber: newContacts[0],
		},
		{
			Id:          id,
			PhoneNumber: newContacts[1],
		},
	}

	mockDb.EXPECT().Insert(&myContacts).Return(nil)

	ctx := context.WithValue(context.Background(), "auth_id", id)
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

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	originalUsername := "UserName"
	username := strings.ToLower(originalUsername)
	findUsername, err := controller.Handler("/username")
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().UsernameIsExist(username).Return(true, nil)

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

	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

	originalUsername := "UserName"
	username := strings.ToLower(originalUsername)
	findUsername, err := controller.Handler("/username")
	if err != nil {
		t.Fatal(err)
	}

	mockDb.EXPECT().UsernameIsExist(username).Return(false, nil)

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
	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

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

	profile := &db.Profile{
		Id:          "idOne",
		Name:        "nameOne",
		Username:    "usernameOne",
		PhoneNumber: "phoneNumber",
		Email:       "email",
		IsMigratory: false,
		AvatarExt:   "png",
		LastUpdate:  123,
	}
	mockDb.EXPECT().ProfileById(id).Return(profile, nil)
	newProfile := *profile
	newProfile.LastUpdate = timestamp.Unix()
	newProfile.AvatarExt = "jpeg"
	mockDb.EXPECT().UpdateByPK(&newProfile).Return(nil)

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
	mockDb := mocks.NewMockDB(ctrl)
	controller := NewController(mockDb)

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

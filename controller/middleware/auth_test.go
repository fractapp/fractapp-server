package middleware

import (
	"context"
	"crypto/sha256"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/mocks"
	"fractapp-server/utils"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/go-chi/jwtauth"

	"github.com/golang/mock/gomock"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var privKey = hexutil.MustDecode("0x507e8ae3b891eefbf35fcd5cac8acb6fc76c5af21e285d5bb43939baa25f5f67")
var pubKeyStr = "0x9af3e86cb6ab6f03de7f5f6fc7874a785a5c15fedc022898e13c4532ccb7bf5f"
var pubKeyBytes = hexutil.MustDecode(pubKeyStr)
var prefix = "It is my fractapp rq:"
var rqBody = "Json Rq"
var tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
var pubKeyHash = sha256.Sum256(pubKeyBytes)
var authId = hexutil.Encode(pubKeyHash[:])[2:]

func TestPubKeyAuthPositive(t *testing.T) {
	var privKeyBytes [32]byte
	copy(privKeyBytes[:], privKey)

	time := strconv.FormatInt(time.Now().Unix(), 10)
	sig, err := utils.Sign(privKeyBytes, []byte(prefix+rqBody+time))
	if err != nil {
		t.Fatal(err.Error())
	}

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode(sig)},
		},
		Body: ioutil.NopCloser(strings.NewReader(rqBody)),
	}

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	id, err := authMiddleware.authWithPubKey(rq)
	if err != nil {
		t.Fatal(err.Error())
	}

	if id != authId {
		t.Fatal("auth returned invalid pubKey")
	}
}
func TestPubKeyAuthInvalidSign(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var sign [64]byte
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(sign[:])

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{timestamp},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode(sign[:])},
		},
		Body: ioutil.NopCloser(strings.NewReader(rqBody)),
	}

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	if _, err := authMiddleware.authWithPubKey(rq); err != InvalidAuthErr {
		t.Fatal()
	}
}
func TestPubKeyAuthInvalidSyntaxTimestamp(t *testing.T) {
	time := "asd"

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode([]byte{})},
		},
		Body: ioutil.NopCloser(strings.NewReader(rqBody)),
	}

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	_, err := authMiddleware.authWithPubKey(rq)

	_, strErr := strconv.ParseInt(time, 10, 64)
	if err.Error() != strErr.Error() {
		t.Fatal(err.Error())
	}
}
func TestPubKeyAuthInvalidTimestamp(t *testing.T) {
	time := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode([]byte{})},
		},
		Body: ioutil.NopCloser(strings.NewReader(rqBody)),
	}

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	_, err := authMiddleware.authWithPubKey(rq)

	if err != controller.InvalidSignTimeErr {
		t.Fatal(err.Error())
	}
}

func TestPubKeyHandlerPositive(t *testing.T) {
	var privKeyBytes [32]byte
	copy(privKeyBytes[:], privKey)

	time := strconv.FormatInt(time.Now().Unix(), 10)
	sig, err := utils.Sign(privKeyBytes, []byte(prefix+rqBody+time))
	if err != nil {
		t.Fatal(err.Error())
	}

	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})

	r := httptest.NewRequest("POST", "http://127.0.0.1:80", strings.NewReader(rqBody))
	r.Header.Add("Sign-Timestamp", time)
	r.Header.Add("Auth-Key", pubKeyStr)
	r.Header.Add("Sign", hexutil.Encode(sig))

	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	h := authMiddleware.PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !isCalled {
		t.Fatal()
	}

}
func TestPubKeyHandlerInvalidSign(t *testing.T) {
	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})
	time := strconv.FormatInt(time.Now().Unix(), 10)

	r := httptest.NewRequest("POST", "http://127.0.0.1:80", nil)
	r.Header.Add("Sign-Timestamp", time)
	r.Header.Add("Auth-Key", pubKeyStr)
	r.Header.Add("Sign", hexutil.Encode([]byte{}))

	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	h := authMiddleware.PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || isCalled {
		t.Fatal()
	}
}
func TestPubKeyHandlerSyntaxTimestamp(t *testing.T) {
	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})

	time := "asd"

	r := httptest.NewRequest("POST", "http://127.0.0.1:80", nil)
	r.Header.Add("Sign-Timestamp", time)
	r.Header.Add("Auth-Key", pubKeyStr)
	r.Header.Add("Sign", hexutil.Encode([]byte{}))

	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	h := authMiddleware.PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest || isCalled {
		t.Fatal()
	}
}
func TestPubKeyHandlerInvalidTimestamp(t *testing.T) {
	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})
	time := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	r := httptest.NewRequest("POST", "http://127.0.0.1:80", nil)
	r.Header.Add("Sign-Timestamp", time)
	r.Header.Add("Auth-Key", pubKeyStr)
	r.Header.Add("Sign", hexutil.Encode([]byte{}))

	w := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	authMiddleware := New(mocks.NewMockDB(ctrl))

	h := authMiddleware.PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || isCalled {
		t.Fatal()
	}
}

func TestJWTAuthPositive(t *testing.T) {
	token, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": authId})
	if err != nil {
		t.Fatal(err)
	}

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, token), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rq.Header = http.Header{
		"Authorization": []string{"BEARER " + tokenString},
	}

	ctrl := gomock.NewController(t)
	db := mocks.NewMockDB(ctrl)
	authMiddleware := New(db)

	db.EXPECT().
		IdByToken(gomock.Eq(tokenString)).
		Return(authId, nil).AnyTimes()

	id, err := authMiddleware.authWithJwt(rq)
	if err != nil {
		t.Fatal(err.Error())
	}

	if id != authId {
		t.Fatal("auth returned invalid pubKey")
	}
}
func TestJWTAuthInvalidToken(t *testing.T) {
	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, "asdaskdljl"), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	db := mocks.NewMockDB(ctrl)
	authMiddleware := New(db)

	_, err = authMiddleware.authWithJwt(rq)
	assert.Assert(t, err == InvalidAuthErr)
}
func TestJWTAuthNegativeNotExistInDb(t *testing.T) {
	token, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": authId})
	if err != nil {
		t.Fatal(err)
	}

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, token), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rq.Header = http.Header{
		"Authorization": []string{"BEARER " + tokenString},
	}

	ctrl := gomock.NewController(t)
	mockDb := mocks.NewMockDB(ctrl)
	authMiddleware := New(mockDb)

	mockDb.EXPECT().
		IdByToken(gomock.Eq(tokenString)).
		Return("", db.ErrNoRows).AnyTimes()

	_, err = authMiddleware.authWithJwt(rq)
	assert.Assert(t, err == InvalidAuthErr)
}
func TestJWTAuthNegativeInvalidClaims(t *testing.T) {
	token, tokenString, err := tokenAuth.Encode(map[string]interface{}{"notId": authId})
	if err != nil {
		t.Fatal(err)
	}

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, token), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rq.Header = http.Header{
		"Authorization": []string{"BEARER " + tokenString},
	}

	ctrl := gomock.NewController(t)
	db := mocks.NewMockDB(ctrl)
	authMiddleware := New(db)

	db.EXPECT().
		IdByToken(gomock.Eq(tokenString)).
		Return(authId, nil).AnyTimes()

	_, err = authMiddleware.authWithJwt(rq)
	assert.Assert(t, err == InvalidAuthErr)
}

func TestJWTHandlerPositive(t *testing.T) {
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Context().Value(AuthIdKey), authId)
	})

	token, tokenString, err := tokenAuth.Encode(map[string]interface{}{"id": authId})
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	db := mocks.NewMockDB(ctrl)
	authMiddleware := New(db)

	db.EXPECT().
		IdByToken(gomock.Eq(tokenString)).
		Return(authId, nil).AnyTimes()

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, token), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rq.Header = http.Header{
		"Authorization": []string{"BEARER " + tokenString},
	}

	w := httptest.NewRecorder()
	h := authMiddleware.JWTAuth(nH)

	h.ServeHTTP(w, rq)
	assert.Equal(t, w.Code, http.StatusOK)
}
func TestJWTHandlerInvalidAuth(t *testing.T) {
	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})

	ctrl := gomock.NewController(t)
	db := mocks.NewMockDB(ctrl)
	authMiddleware := New(db)

	rq, err := http.NewRequestWithContext(context.WithValue(context.Background(), jwtauth.TokenCtxKey, "asdasd"), "POST", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h := authMiddleware.JWTAuth(nH)

	h.ServeHTTP(w, rq)
	assert.Equal(t, w.Code, http.StatusUnauthorized)
	assert.Equal(t, isCalled, false)
}

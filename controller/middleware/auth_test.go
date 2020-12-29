package middleware

import (
	"crypto/sha256"
	"fractapp-server/controller"
	"fractapp-server/utils"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var privKey = hexutil.MustDecode("0x507e8ae3b891eefbf35fcd5cac8acb6fc76c5af21e285d5bb43939baa25f5f67")
var pubKeyStr = "0x9af3e86cb6ab6f03de7f5f6fc7874a785a5c15fedc022898e13c4532ccb7bf5f"
var pubKeyBytes = hexutil.MustDecode(pubKeyStr)
var prefix = "It is my fractapp rq:"
var rq = "Json Rq"

func TestAuthPositive(t *testing.T) {
	var privKeyBytes [32]byte
	copy(privKeyBytes[:], privKey)

	time := strconv.FormatInt(time.Now().Unix(), 10)
	sig, err := utils.Sign(privKeyBytes, []byte(prefix+rq+time))
	if err != nil {
		t.Fatal(err.Error())
	}

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode(sig)},
		},
		Body: ioutil.NopCloser(strings.NewReader(rq)),
	}

	id, err := auth(rq)
	if err != nil {
		t.Fatal(err.Error())
	}

	hash := sha256.Sum256(pubKeyBytes)
	if id != hexutil.Encode(hash[:]) {
		t.Fatal("auth returned invalid pubKey")
	}
}
func TestAuthInvalidSign(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var sign [64]byte
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(sign[:])

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{timestamp},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode(sign[:])},
		},
		Body: ioutil.NopCloser(strings.NewReader(rq)),
	}

	if _, err := auth(rq); err != utils.InvalidSignErr {
		t.Fatal()
	}
}
func TestAuthInvalidSyntaxTimestamp(t *testing.T) {
	time := "asd"

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode([]byte{})},
		},
		Body: ioutil.NopCloser(strings.NewReader(rq)),
	}

	_, err := auth(rq)

	_, strErr := strconv.ParseInt(time, 10, 64)
	if err.Error() != strErr.Error() {
		t.Fatal(err.Error())
	}
}
func TestAuthInvalidTimestamp(t *testing.T) {
	time := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	rq := &http.Request{
		Header: http.Header{
			"Sign-Timestamp": []string{time},
			"Auth-Key":       []string{pubKeyStr},
			"Sign":           []string{hexutil.Encode([]byte{})},
		},
		Body: ioutil.NopCloser(strings.NewReader(rq)),
	}

	_, err := auth(rq)

	if err != controller.InvalidSignTimeErr {
		t.Fatal(err.Error())
	}
}

func TestHandlerPositive(t *testing.T) {
	var privKeyBytes [32]byte
	copy(privKeyBytes[:], privKey)

	time := strconv.FormatInt(time.Now().Unix(), 10)
	sig, err := utils.Sign(privKeyBytes, []byte(prefix+rq+time))
	if err != nil {
		t.Fatal(err.Error())
	}

	isCalled := false
	nH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
	})

	r := httptest.NewRequest("POST", "http://127.0.0.1:80", strings.NewReader(rq))
	r.Header.Add("Sign-Timestamp", time)
	r.Header.Add("Auth-Key", pubKeyStr)
	r.Header.Add("Sign", hexutil.Encode(sig))

	w := httptest.NewRecorder()

	h := PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !isCalled {
		t.Fatal()
	}

}
func TestHandlerInvalidSign(t *testing.T) {
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

	h := PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || isCalled {
		t.Fatal()
	}
}
func TestHandlerSyntaxTimestamp(t *testing.T) {
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

	h := PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest || isCalled {
		t.Fatal()
	}
}
func TestHandlerInvalidTimestamp(t *testing.T) {
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

	h := PubKeyAuth(nH)

	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || isCalled {
		t.Fatal()
	}
}

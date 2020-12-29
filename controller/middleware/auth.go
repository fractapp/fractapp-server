package middleware

import (
	"context"
	"crypto/sha256"
	"fractapp-server/controller"
	"fractapp-server/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Header string

const (
	AuthMsg          = "It is my fractapp rq:"
	AuthIdKey string = "auth_id"

	SignTimestamp Header = "Sign-Timestamp"
	Sign          Header = "Sign"
	AuthPubKey    Header = "Auth-Key"
)

func PubKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := auth(r)
		if err == controller.InvalidSignTimeErr || err == utils.InvalidSignErr {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, controller.InvalidAuthErr.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), AuthIdKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func auth(r *http.Request) (string, error) {
	strTimestamp := r.Header.Get(string(SignTimestamp))
	hexPubKey := r.Header.Get(string(AuthPubKey))
	hexSign := r.Header.Get(string(Sign))

	timestamp, err := strconv.ParseInt(strTimestamp, 10, 64)
	if err != nil {
		return "", err
	}

	rqTime := time.Unix(timestamp, 0)
	if rqTime.Add(controller.SignTimeout).Before(time.Now()) {
		return "", controller.InvalidSignTimeErr
	}

	rq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	msg := AuthMsg + string(rq) + strconv.FormatInt(rqTime.Unix(), 10)
	pubKey, err := utils.ParsePubKey(hexPubKey)
	if err != nil {
		return "", err
	}

	if err := utils.Verify(pubKey, msg, hexSign); err != nil {
		return "", err
	}

	hash := sha256.Sum256(pubKey[:])
	return hexutil.Encode(hash[:]), nil
}

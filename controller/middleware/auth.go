package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/utils"

	"github.com/lestrrat-go/jwx/jwt"

	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Header string

const (
	AuthMsg             = "It is my fractapp rq:"
	AuthIdKey    string = "auth_id"
	ProfileIdKey string = "profile_id"

	SignTimestamp Header = "Sign-Timestamp"
	Sign          Header = "Sign"
	AuthPubKey    Header = "Auth-Key"
)

var (
	InvalidAuthErr = errors.New("invalid auth")
)

type AuthMiddleware struct {
	db db.DB
}

func New(db db.DB) *AuthMiddleware {
	return &AuthMiddleware{
		db: db,
	}
}

func AuthId(r *http.Request) string {
	return r.Context().Value(AuthIdKey).(string)
}

func ProfileId(r *http.Request) db.ID {
	return r.Context().Value(ProfileIdKey).(db.ID)
}

func (a *AuthMiddleware) PubKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := a.authWithPubKey(r)
		if err == controller.InvalidSignTimeErr || err == InvalidAuthErr {
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
func (a *AuthMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authId, profileId, err := a.AuthWithJwt(r, jwtauth.TokenFromHeader)
		if err == InvalidAuthErr {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, controller.InvalidAuthErr.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), AuthIdKey, authId)
		ctx = context.WithValue(ctx, ProfileIdKey, profileId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *AuthMiddleware) authWithPubKey(r *http.Request) (string, error) {
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

	defer r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(rq))

	msg := AuthMsg + string(rq) + strconv.FormatInt(rqTime.Unix(), 10)
	pubKey, err := utils.ParsePubKey(hexPubKey)
	if err != nil {
		return "", err
	}

	if err := utils.Verify(pubKey, msg, hexSign); err != nil {
		return "", InvalidAuthErr
	}

	hash := sha256.Sum256(pubKey[:])
	return hexutil.Encode(hash[:])[2:], nil
}
func (a *AuthMiddleware) AuthWithJwt(r *http.Request, findTokenFns func(r *http.Request) string) (string, db.ID, error) {
	token, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", db.ID{}, err
	}
	if token == nil || jwt.Validate(token) != nil {
		return "", db.ID{}, InvalidAuthErr
	}

	tokenDb, err := a.db.TokenByValue(findTokenFns(r))
	if err != nil {
		return "", db.ID{}, InvalidAuthErr
	}

	p, err := a.db.ProfileById(tokenDb.ProfileId)
	if err != nil {
		return "", db.ID{}, InvalidAuthErr
	}

	if p.AuthId != claims["id"] {
		return "", db.ID{}, InvalidAuthErr
	}

	return p.AuthId, p.Id, nil
}

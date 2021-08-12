package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/notification"
	"fractapp-server/types"
	"fractapp-server/utils"
	"fractapp-server/validators"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"
)

const (
	SignAddressMsg       = "It is my auth key for fractapp:"
	CodeSendTimeout      = 3 * time.Minute
	CodeTimeout          = 10 * time.Minute
	ResetTimeout         = 1 * time.Hour
	MaxSendCount         = 5
	MaxWrongCodeAttempts = 3

	SendCodeRoute = "/sendCode"
	SignInRoute   = "/signIn"
)

var (
	InvalidSendTimeoutErr = errors.New("expect timeout for new code sending")

	AddressExistErr            = errors.New("address exist")
	AccountExistErr            = errors.New("account exist")
	InvalidCodeErr             = errors.New("invalid confirm code")
	InvalidNumberOfAttemptsErr = errors.New("invalid number of attempts")
	CodeUsedErr                = errors.New("code used")
	CodeExpiredErr             = errors.New("code expired")
)

type Controller struct {
	db          db.DB
	notificator map[notification.NotificatorType]notification.Notificator
	jwtauth     *jwtauth.JWTAuth
}

func NewController(db db.DB, sms notification.Notificator,
	email notification.Notificator, jwtauth *jwtauth.JWTAuth) *Controller {
	return &Controller{
		db: db,
		notificator: map[notification.NotificatorType]notification.Notificator{
			notification.Email: email,
			notification.SMS:   sms,
		},
		jwtauth: jwtauth,
	}
}

func (c *Controller) MainRoute() string {
	return "/auth"
}
func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case SendCodeRoute:
		return c.sendCode, nil
	case SignInRoute:
		return c.signIn, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case notification.InvalidEmailErr:
		fallthrough
	case InvalidCodeErr:
		fallthrough
	case notification.InvalidPhoneNumberErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	case InvalidSendTimeoutErr:
		http.Error(w, err.Error(), http.StatusAccepted)
	case CodeExpiredErr:
		fallthrough
	case CodeUsedErr:
		fallthrough
	case InvalidNumberOfAttemptsErr:
		http.Error(w, err.Error(), http.StatusTooManyRequests)
	case AddressExistErr:
		fallthrough
	case AccountExistErr:
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// sendCode godoc
// @Summary Send code
// @Description send auth code to email/phone
// @ID send-auth-code
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Param rq body SendCodeRq true "Send code rq"
// @Success 200
// @Failure 404 {string} string notification.InvalidPhoneNumberErr
// @Failure 404 {string} string notification.InvalidEmailErr:
// @Failure 202 {string} string InvalidSendTimeoutErr
// @Failure 400 {string} string
// @Router /auth/sendCode [post]
func (c *Controller) sendCode(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	rq := SendCodeRq{}
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	log.Printf("Type: %d Value: %s\n", rq.Type, rq.Value)

	rq.Value = c.notificator[rq.Type].Format(rq.Value)
	if err := c.notificator[rq.Type].Validate(rq.Value); err != nil {
		return err
	}

	auth, err := c.db.AuthByValue(rq.Value, rq.Type)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	now := time.Now()

	if err == db.ErrNoRows || auth == nil {
		auth = &db.Auth{
			Id:      db.NewId(),
			Value:   rq.Value,
			Type:    rq.Type,
			IsValid: true,
		}
	} else {
		if now.Before(time.Unix(auth.Timestamp, 0).Add(CodeSendTimeout)) ||
			(auth.Count >= MaxSendCount && now.Before(time.Unix(auth.Timestamp, 0).Add(ResetTimeout))) {

			return InvalidSendTimeoutErr
		}
	}

	if now.After(time.Unix(auth.Timestamp, 0).Add(ResetTimeout)) {
		auth.Count = 0
	}

	code := generateCode()

	auth.Code = code
	auth.Timestamp = now.Unix()
	auth.Count++
	auth.Attempts = 0
	auth.IsValid = true

	if err == db.ErrNoRows {
		err = c.db.Insert(auth)
	} else {
		err = c.db.UpdateByPK(auth.Id, auth)
	}

	if err != nil {
		return err
	}

	if err := c.notificator[rq.Type].SendCode(rq.Value, code); err != nil {
		return err
	}

	return nil
}

// signIn godoc
// @Summary Sign in
// @Description sign in to fractapp account
// @ID signIn
// @Security AuthWithPubKey-SignTimestamp
// @Security AuthWithPubKey-Sign
// @Security AuthWithPubKey-Auth-Key
// @Tags Authorization
// @Accept  json
// @Produce json
// @Param rq body ConfirmAuthRq true "Confirm auth rq"
// @Success 200 {object} TokenRs
// @Failure 429 {string} string CodeExpiredErr
// @Failure 429 {string} string CodeUsedErr
// @Failure 429 {string} string InvalidNumberOfAttemptsErr
// @Failure 403 {string} string AddressExistErr
// @Failure 403 {string} string AccountExistErr
// @Failure 400 {string} string
// @Router /auth/signIn [post]
func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	rq := &ConfirmAuthRq{}
	err = json.Unmarshal(b, rq)
	if err != nil {
		return err
	}

	if rq.Type != notification.CryptoAddress {
		rq.Value = c.notificator[rq.Type].Format(rq.Value)
		if err := c.notificator[rq.Type].Validate(rq.Value); err != nil {
			return err
		}

		//check confirm code
		if err := c.confirm(rq.Value, rq.Type, rq.Code); err != nil {
			return err
		}
	}

	id := middleware.AuthId(r)
	profile, err := c.db.ProfileByAuthId(id)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	existProfile := &db.Profile{}
	switch rq.Type {
	case notification.Email:
		existProfile, err = c.db.ProfileByEmail(rq.Value)
	case notification.SMS:
		existProfile, err = c.db.ProfileByPhoneNumber(rq.Value)
	case notification.CryptoAddress:
		existProfile, err = c.db.ProfileByAuthId(id)
	}

	if err != nil && err != db.ErrNoRows {
		return err
	}
	if err != db.ErrNoRows && existProfile != nil && existProfile.AuthId != id {
		return AccountExistErr
	}

	//check time and sign
	strTimestamp := r.Header.Get(string(middleware.SignTimestamp))
	timestamp, err := strconv.ParseInt(strTimestamp, 10, 64)
	if err != nil {
		return err
	}

	rqTime := time.Unix(timestamp, 0)
	now := time.Now()
	if now.After(rqTime.Add(controller.SignTimeout)) || now.After(rqTime.Add(1*time.Minute)) {
		return controller.InvalidSignTimeErr
	}

	err = c.checkAddresses(rq, r.Header.Get(string(middleware.AuthPubKey)), rqTime, profile)
	if err != nil {
		return err
	}

	// if user was registered that check addresses
	if profile != nil {
		switch rq.Type {
		case notification.Email:
			profile.Email = rq.Value
		case notification.SMS:
			profile.PhoneNumber = rq.Value
		case notification.CryptoAddress:
		}
		err = c.db.UpdateByPK(profile.Id, profile)
		if err != nil {
			return err
		}
	} else {
		addresses := make(map[types.Network]db.Address)
		for network, v := range rq.Addresses {
			addresses[network] = db.Address{
				Address: v.Address,
			}
		}

		total, err := c.db.ProfilesCount()
		if err != nil {
			return err
		}
		username := fmt.Sprintf("%s%d", validators.UsernamePrefix, total)

		profile = &db.Profile{
			Id:        db.NewId(),
			AuthId:    id,
			Username:  username,
			Addresses: addresses,
		}

		switch rq.Type {
		case notification.Email:
			profile.Email = rq.Value
		case notification.SMS:
			profile.PhoneNumber = rq.Value
		case notification.CryptoAddress:
		}

		if err := c.db.Insert(profile); err != nil {
			return err
		}
	}

	_, tokenString, err := c.jwtauth.Encode(map[string]interface{}{"id": id, "timestamp": now.Unix()})
	if err != nil {
		return err
	}

	dbToken, err := c.db.TokenByProfileId(profile.Id)
	if err == db.ErrNoRows {
		err = c.db.Insert(&db.Token{Id: db.NewId(), Token: tokenString, ProfileId: profile.Id})
	} else {
		dbToken.Token = tokenString
		err = c.db.UpdateByPK(dbToken.Id, dbToken)
	}
	if err != nil {
		return err
	}

	tokenRs := &TokenRs{Token: tokenString}
	rsByte, err := json.Marshal(tokenRs)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) confirm(value string, codeType notification.NotificatorType, code string) error {
	auth, err := c.db.AuthByValue(value, codeType)
	if err != nil {
		return err
	}

	if auth.Attempts >= MaxWrongCodeAttempts {
		return InvalidNumberOfAttemptsErr
	}

	if !auth.IsValid {
		return CodeUsedErr
	}

	if time.Unix(auth.Timestamp, 0).Add(CodeTimeout).Before(time.Now()) {
		return CodeExpiredErr
	}

	if auth.Code != code {
		auth.Attempts++

		if err := c.db.UpdateByPK(auth.Id, auth); err != nil {
			return err
		}

		return InvalidCodeErr
	}
	auth.IsValid = false

	c.db.UpdateByPK(auth.Id, auth)

	return nil
}
func (c *Controller) checkAddresses(rq *ConfirmAuthRq, authPubKey string, rqTime time.Time, profile *db.Profile) error {
	if len(rq.Addresses) != 2 {
		return controller.InvalidRqErr
	}
	if _, ok := rq.Addresses[types.Polkadot]; !ok {
		return controller.InvalidRqErr
	}
	if _, ok := rq.Addresses[types.Kusama]; !ok {
		return controller.InvalidRqErr
	}

	msg := SignAddressMsg + authPubKey + strconv.FormatInt(rqTime.Unix(), 10)
	for network, v := range rq.Addresses {
		pubKey, err := utils.ParsePubKey(v.PubKey)
		if err != nil {
			return err
		}

		if network.Address(pubKey[:]) != v.Address {
			return controller.InvalidAddressErr
		}
		if err := utils.Verify(pubKey, msg, v.Sign); err != nil {
			return err
		}

		if profile != nil {
			continue
		}

		isAddressExist := true
		_, err = c.db.ProfileByAddress(network, v.Address)
		if err != db.ErrNoRows && err != nil {
			return err
		}

		if err == db.ErrNoRows {
			isAddressExist = false
		} else {
			isAddressExist = true
		}

		if isAddressExist {
			return AddressExistErr
		}
	}

	if profile != nil {
		pDb, err := c.db.ProfileByAuthId(profile.AuthId)
		if err != nil {
			return err
		}

		for k, v := range pDb.Addresses {
			if rq.Addresses[k].Address != v.Address {
				return AccountExistErr
			}
		}
	}
	return nil
}

package auth

import (
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/notification"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"

	"github.com/go-pg/pg/v10"
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
// @Tags auth
// @Accept  json
// @Produce  json
// @Param rq body SendCodeRq true "Send code rq"
// @Success 200
// @Failure 404 string notification.InvalidPhoneNumberErr
// @Failure 404 string notification.InvalidEmailErr:
// @Failure 202 string InvalidSendTimeoutErr
// @Failure 400
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

	auth, err := c.db.AuthByValue(rq.Value, rq.Type, rq.CheckType)
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	now := time.Now()

	if err == pg.ErrNoRows || auth == nil {
		auth = &db.Auth{
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

	auth.CheckType = rq.CheckType
	auth.Code = code
	auth.Timestamp = now.Unix()
	auth.Count++
	auth.Attempts = 0
	auth.IsValid = true

	if err == pg.ErrNoRows {
		err = c.db.Insert(auth)
	} else {
		err = c.db.UpdateByPK(auth)
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
// @Tags auth
// @Accept  json
// @Produce  json
// @Param rq body ConfirmAuthRq true "Confirm auth rq"
// @Success 200
// @Failure 429 string CodeExpiredErr
// @Failure 429 string CodeUsedErr
// @Failure 429 string InvalidNumberOfAttemptsErr
// @Failure 403 string AddressExistErr
// @Failure 403 string AccountExistErr
// @Failure 400
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

	rq.Value = c.notificator[rq.Type].Format(rq.Value)

	//check confirm code
	if err := c.confirm(rq.Value, rq.Type, notification.Auth, rq.Code); err != nil {
		return err
	}

	id := middleware.AuthId(r)
	profile, err := c.db.ProfileById(id)
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	if profile != nil && profile.Id != id {
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
		}
		err = c.db.UpdateByPK(profile)
		if err != nil {
			return err
		}
	} else {
		var addresses []*db.Address
		for network, v := range rq.Addresses {
			addresses = append(addresses, &db.Address{
				Id:      id,
				Address: v.Address,
				Network: network,
			})
		}

		profile = &db.Profile{
			Id:          id,
			IsMigratory: false,
		}

		switch rq.Type {
		case notification.Email:
			profile.Email = rq.Value
		case notification.SMS:
			profile.PhoneNumber = rq.Value
		}

		if err := c.db.CreateProfile(r.Context(), profile, addresses); err != nil {
			return err
		}
	}

	_, tokenString, err := c.jwtauth.Encode(map[string]interface{}{"id": id, "timestamp": now.Unix()})
	if err != nil {
		return err
	}

	token := &TokenRs{Token: tokenString}
	rsByte, err := json.Marshal(token)
	if err != nil {
		return err
	}

	_, err = c.db.TokenById(id)
	if err == db.ErrNoRows {
		err = c.db.Insert(&db.Token{Token: tokenString, Id: id})
	} else {
		err = c.db.Update(&db.Token{Token: tokenString, Id: id}, "id = ?", id)
	}
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) confirm(value string, codeType notification.NotificatorType, checkType notification.CheckType, code string) error {
	auth, err := c.db.AuthByValue(value, codeType, checkType)
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

		if err := c.db.UpdateByPK(auth); err != nil {
			return err
		}

		return InvalidCodeErr
	}
	auth.IsValid = false

	c.db.UpdateByPK(auth)

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

		addressExist, err := c.db.AddressIsExist(v.Address)
		if err != nil {
			return err
		}

		if addressExist {
			return AddressExistErr
		}
	}

	if profile != nil {
		addresses, err := c.db.AddressesById(profile.Id)
		if err != nil {
			return err
		}
		for _, v := range addresses {
			if rq.Addresses[v.Network].Address != v.Address {
				return AccountExistErr
			}
		}
	}
	return nil
}

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/email"
	"fractapp-server/twilio"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"

	"github.com/go-pg/pg/v10"
)

const (
	SignAddressMsg   = "It is my auth key for fractapp:"
	SMSMsg           = "Your code for Fractapp: %s"
	SMSTimeout       = 3 * time.Minute
	ResetCountSMS    = 1 * time.Hour
	MaxSMSCount      = 5
	MaxWrongCodeSend = 3

	SendCodeRoute = "/sendCode"
	SignInRoute   = "/signIn"
)

var (
	InvalidSendSMSTimeoutErr = errors.New("expect timeout for new SMS sending")

	AddressExistErr         = errors.New("address exist")
	AccountExistErr         = errors.New("account exist")
	InvalidCodeErr          = errors.New("invalid confirm code")
	InvalidNumberOfAttempts = errors.New("invalid number of attempts")
)

type Controller struct {
	db          db.DB
	twilio      *twilio.Api
	emailClient *email.Client
	jwtauth     *jwtauth.JWTAuth
}

func NewController(db db.DB, fromNumber string, accountSid string, authToken string,
	emailClient *email.Client, jwtauth *jwtauth.JWTAuth) *Controller {
	return &Controller{
		db:          db,
		twilio:      twilio.NewApi(fromNumber, accountSid, authToken),
		emailClient: emailClient,
		jwtauth:     jwtauth,
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
	case email.InvalidEmailErr:
		fallthrough
	case InvalidCodeErr:
		fallthrough
	case twilio.InvalidPhoneNumberErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	case InvalidSendSMSTimeoutErr:
		http.Error(w, err.Error(), http.StatusAccepted)
	case InvalidNumberOfAttempts:
		http.Error(w, err.Error(), http.StatusTooManyRequests)
	case AddressExistErr:
		fallthrough
	case AccountExistErr:
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

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
	switch rq.Type {
	case types.PhoneNumberCode:
		rq.Value = c.twilio.FormatNumber(rq.Value)
		if err := c.twilio.ValidatePhoneNumber(rq.Value); err != nil {
			return err
		}
	case types.EmailCode:
		rq.Value = c.emailClient.FormatEmail(rq.Value)
		if err := c.emailClient.ValidateEmail(rq.Value); err != nil {
			return err
		}
	}

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	codeInt := generator.Intn(999999)
	code := fmt.Sprintf("%d", codeInt)
	if len(code) < 6 {
		code = fmt.Sprintf("%06s", code)
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
		if now.Before(time.Unix(auth.Timestamp, 0).Add(SMSTimeout)) {
			return InvalidSendSMSTimeoutErr
		}
		if now.Before(time.Unix(auth.Timestamp, 0).Add(ResetCountSMS)) && auth.Count >= MaxSMSCount {
			return InvalidSendSMSTimeoutErr
		}
	}

	if now.After(time.Unix(auth.Timestamp, 0).Add(ResetCountSMS)) {
		auth.Count = 0
	}

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

	switch rq.Type {
	case types.PhoneNumberCode:
		if err := c.twilio.SendSMS(rq.Value, SMSMsg, code); err != nil {
			return err
		}
	case types.EmailCode:
		if err := c.emailClient.SendCode(rq.Value, code); err != nil {
			return err
		}
	}

	return nil
}
func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	rq := ConfirmRegRq{}
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	switch rq.Type {
	case types.PhoneNumberCode:
		rq.Value = c.twilio.FormatNumber(rq.Value)
	case types.EmailCode:
		rq.Value = c.emailClient.FormatEmail(rq.Value)
	}

	//check confirm code
	if err := c.confirm(rq.Value, rq.Type, types.Auth, rq.Code); err != nil {
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
	if rqTime.After(time.Now().Add(controller.SignTimeout)) {
		return controller.InvalidSignTimeErr
	}

	if len(rq.Addresses) != 2 {
		return controller.InvalidRqErr
	}
	if _, ok := rq.Addresses[types.Polkadot]; !ok {
		return controller.InvalidRqErr
	}
	if _, ok := rq.Addresses[types.Kusama]; !ok {
		return controller.InvalidRqErr
	}

	msg := SignAddressMsg + r.Header.Get(string(middleware.AuthPubKey)) + strconv.FormatInt(rqTime.Unix(), 10)
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

	// if user was registered that check addresses
	if profile != nil {
		addresses, err := c.db.AddressesById(id)
		if err != nil {
			return err
		}
		for _, v := range addresses {
			if rq.Addresses[v.Network].Address != v.Address {
				return AccountExistErr
			}
		}
	}

	if profile == nil {
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
		case types.EmailCode:
			profile.Email = rq.Value
		case types.PhoneNumberCode:
			profile.PhoneNumber = rq.Value
		}

		if err := c.db.CreateProfile(r.Context(), profile, addresses); err != nil {
			return err
		}
	} else {
		switch rq.Type {
		case types.EmailCode:
			profile.Email = rq.Value
		case types.PhoneNumberCode:
			profile.PhoneNumber = rq.Value
		}
		err = c.db.UpdateByPK(profile)
		if err != nil {
			return err
		}
	}
	_, tokenString, err := c.jwtauth.Encode(map[string]interface{}{"id": id})
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
func (c *Controller) confirm(value string, codeType types.CodeType, checkType types.CheckType, code string) error {
	auth, err := c.db.AuthByValue(value, codeType, checkType)
	if err != nil {
		return err
	}

	if auth.Attempts >= MaxWrongCodeSend {
		return InvalidNumberOfAttempts
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

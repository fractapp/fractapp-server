package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/twilio"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

type Route string

const (
	SignAddressMsg   = "It is my auth key for fractapp:"
	SMSMsg           = "Your code for Fractapp: %s"
	SMSTimeout       = 1 * time.Minute
	ResetCountSMS    = 1 * time.Hour
	MaxSMSCount      = 5
	MaxWrongCodeSend = 3

	MaxUsernameLength = 30
	MaxNameLength     = 40

	//TODO take more symbols
	InvalidSym = "@?.,&^%$#@!^&*()-+=:''`?.,"

	Auth          Route = "auth"
	ConfirmAuth   Route = "confirm"
	UpdateProfile Route = "updateProfile"
	Username      Route = "username"
)

var (
	InvalidSendSMSTimeoutErr = errors.New("expect timeout for new SMS sending")
	AccountExistErr          = errors.New("account exist")
	AddressExistErr          = errors.New("address exist")
	InvalidCodeErr           = errors.New("invalid confirm code")
	InvalidNumberOfAttempts  = errors.New("invalid number of attempts")

	UsernameIsExistErr  = errors.New("username is exist")
	UsernameNotFoundErr = errors.New("username not found")
	InvalidPropertyErr  = errors.New("property has invalid symbols or length")
)

type Controller struct {
	db     *pg.DB
	twilio *twilio.Api
}

func NewController(db *pg.DB, fromNumber string, accountSid string, authToken string) *Controller {
	return &Controller{
		db:     db,
		twilio: twilio.NewApi(fromNumber, accountSid, authToken),
	}
}

func (c *Controller) Route(route Route) func(w http.ResponseWriter, r *http.Request) {
	var f func(r *http.Request) error
	switch route {
	case Auth:
		f = c.auth
	case ConfirmAuth:
		f = c.confirmAuth
	case UpdateProfile:
		f = c.updateProfile
	case Username:
		f = c.findUsername
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(r); err != nil {
			log.Printf("Http error: %s \n", err.Error())

			c.returnErr(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (c *Controller) returnErr(err error, w http.ResponseWriter) {
	switch err {
	case twilio.InvalidPhoneNumberErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	case InvalidSendSMSTimeoutErr:
		http.Error(w, err.Error(), http.StatusAccepted)
	case UsernameNotFoundErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *Controller) auth(r *http.Request) error {
	phoneNumber := "+" + strings.ReplaceAll(r.URL.Query()["number"][0], " ", "")
	log.Printf("Phone number: %s\n", phoneNumber)

	if err := c.twilio.ValidatePhoneNumber(phoneNumber); err != nil {
		return err
	}

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	codeInt := generator.Intn(999999)
	code := fmt.Sprintf("%d", codeInt)
	if len(code) < 6 {
		code = fmt.Sprintf("%06s", code)
	}

	auth := &db.Auth{}
	err := c.db.Model(auth).Where("phone_number = ?", phoneNumber).Select()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	now := time.Now()

	if err == pg.ErrNoRows {
		auth = &db.Auth{
			PhoneNumber: phoneNumber,
			Code:        code,
			Count:       0,
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
	auth.Timestamp = time.Now().Unix()
	auth.Count++

	if err == pg.ErrNoRows {
		_, err = c.db.Model(auth).Insert()
	} else {
		_, err = c.db.Model(auth).
			Where("phone_number = ?", auth.PhoneNumber).
			Update()
	}

	if err := c.twilio.SendSMS(phoneNumber, SMSMsg, code); err != nil {
		return err
	}

	return nil
}
func (c *Controller) confirmAuth(r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := r.Context().Value(middleware.AuthIdKey).(string)

	exist, err := c.db.Model(&db.Profile{}).Where("id = ?", id).Exists()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	if exist {
		return AccountExistErr
	}

	rq := ConfirmRegRq{}
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	strTimestamp := r.Header.Get(string(middleware.SignTimestamp))
	timestamp, err := strconv.ParseInt(strTimestamp, 10, 64)
	if err != nil {
		return err
	}

	rqTime := time.Unix(timestamp, 0)
	if rqTime.After(time.Now().Add(controller.SignTimeout)) {
		return controller.InvalidSignTimeErr
	}

	if len(rq.Addresses) < 2 ||
		rq.Addresses[0].Network != types.Polkadot ||
		rq.Addresses[1].Network != types.Kusama {
		return controller.InvalidRqErr
	}

	msg := SignAddressMsg + r.Header.Get(string(middleware.AuthPubKey)) + strconv.FormatInt(rqTime.Unix(), 10)

	for _, v := range rq.Addresses {
		pubKey, err := utils.ParsePubKey(v.PubKey)
		if err != nil {
			return err
		}

		if v.Network.Address(pubKey[:]) != v.Address {
			return controller.InvalidAddressErr
		}
		if err := utils.Verify(pubKey, msg, v.Sign); err != nil {
			return err
		}

		addressExist, err := c.db.Model(&db.Address{}).Where("address = ?", v.Address).Exists()
		if err != nil {
			return err
		}

		if addressExist {
			return AddressExistErr
		}
	}

	auth := db.Auth{}
	err = c.db.Model(&auth).Where("phoneNumber = ?", rq.PhoneNumber).Select()
	if err != nil {
		return err
	}

	if auth.Attempts >= MaxWrongCodeSend {
		return InvalidNumberOfAttempts
	}

	if auth.Code != strconv.Itoa(rq.Code) {
		auth.Attempts++

		_, err := c.db.Model(&auth).Where("attempts = ?", auth.Attempts).Update()
		if err != nil {
			return err
		}

		return InvalidCodeErr
	}

	if err := c.db.RunInTransaction(r.Context(), func(tx *pg.Tx) error {
		if _, err = c.db.Model(&db.Profile{
			Id:          id,
			PhoneNumber: rq.PhoneNumber,
		}).Insert(); err != nil {
			return err
		}

		for _, v := range rq.Addresses {
			if _, err := c.db.Model(&db.Address{
				Id:      id,
				Address: v.Address,
				Network: v.Network,
			}).Insert(); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
func (c *Controller) updateProfile(r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := r.Context().Value(middleware.AuthIdKey).(string)

	profile := &db.Profile{}
	err = c.db.Model(profile).Where("id = ?", id).First()
	if err != nil {
		return err
	}

	rq := UpdateProfileRq{}
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	if profile.Username != rq.Username {
		isExist, err := c.usernameIsExist(rq.Username)
		if err != nil {
			return err
		}

		if isExist {
			return UsernameIsExistErr
		}

		profile.Username = rq.Username
	}
	if profile.Name != rq.Name {
		if !isValidName(rq.Name) {
			return InvalidPropertyErr
		}

		profile.Name = rq.Name
	}

	_, err = c.db.Model(profile).Update()
	if err != nil {
		return err
	}

	return nil
}
func (c *Controller) findUsername(r *http.Request) error {
	exist, err := c.usernameIsExist(r.URL.Query()["number"][0])
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	return UsernameNotFoundErr
}
func (c *Controller) usernameIsExist(username string) (bool, error) {
	if !isValidUsername(username) {
		return false, InvalidPropertyErr
	}

	count, err := c.db.Model(&db.Profile{}).Where("username = ?", username).Count()
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

//TODO transfer to any pkg
func isValidUsername(username string) bool {
	if len(username) > MaxUsernameLength {
		return false
	}

	count := strings.Count(" ", username)
	//TODO test
	for _, v := range InvalidSym {
		count += strings.Count(string(v), username)
	}

	return count == 0
}

//TODO  transfer to any pkg
func isValidName(name string) bool {
	if len(name) > MaxNameLength {
		return false
	}

	var count int
	//TODO test
	for _, v := range InvalidSym {
		count += strings.Count(string(v), name)
	}

	return count == 0
}

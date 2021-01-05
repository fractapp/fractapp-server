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
	"fractapp-server/validators"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

const (
	SignAddressMsg   = "It is my auth key for fractapp:"
	SMSMsg           = "Your code for Fractapp: %s"
	SMSTimeout       = 1 * time.Minute
	ResetCountSMS    = 1 * time.Hour
	MaxSMSCount      = 5
	MaxWrongCodeSend = 3

	AuthRoute          = "/auth"
	ConfirmAuthRoute   = "/confirm"
	UpdateProfileRoute = "/updateProfile"
	UsernameRoute      = "/username"
)

var (
	InvalidSendSMSTimeoutErr = errors.New("expect timeout for new SMS sending")
	InvalidSignInErr         = errors.New("invalid sign in")
	AddressExistErr          = errors.New("address exist")
	InvalidCodeErr           = errors.New("invalid confirm code")
	InvalidNumberOfAttempts  = errors.New("invalid number of attempts")

	UsernameIsExistErr  = errors.New("username is exist")
	UsernameNotFoundErr = errors.New("username not found")
	InvalidPropertyErr  = errors.New("property has invalid symbols or length")
)

type Controller struct {
	db     db.DB
	twilio *twilio.Api
}

func NewController(db db.DB, fromNumber string, accountSid string, authToken string) *Controller {
	return &Controller{
		db:     db,
		twilio: twilio.NewApi(fromNumber, accountSid, authToken),
	}
}

func (c *Controller) MainRoute() string {
	return "/profile"
}
func (c *Controller) Handler(route string) (func(r *http.Request) error, error) {
	switch route {
	case AuthRoute:
		return c.auth, nil
	case ConfirmAuthRoute:
		return c.confirmAuth, nil
	case UpdateProfileRoute:
		return c.updateProfile, nil
	case UsernameRoute:
		return c.findUsername, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
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

	auth, err := c.db.AuthByPhoneNumber(phoneNumber)
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	now := time.Now()

	if err == pg.ErrNoRows {
		auth = &db.Auth{
			PhoneNumber: phoneNumber,
			Code:        code,
			Type:        types.PhoneNumberCode,
			CheckType:   types.Auth,
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
		err = c.db.Insert(auth)
	} else {
		err = c.db.UpdateByPK(auth)
	}
	if err != nil {
		return err
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

	id := middleware.AuthId(r)
	profile, err := c.db.ProfileById(id)
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	if c != nil && profile.Id != id {
		return InvalidSignInErr
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

		addressExist, err := c.db.AddressIsExist(v.Address)
		if err != nil {
			return err
		}

		if addressExist {
			return AddressExistErr
		}
	}

	auth, err := c.db.AuthByPhoneNumber(rq.PhoneNumber)
	if err != nil {
		return err
	}

	if auth.Attempts >= MaxWrongCodeSend {
		return InvalidNumberOfAttempts
	}

	if auth.Code != strconv.Itoa(rq.Code) {
		auth.Attempts++

		if err := c.db.UpdateByPK(auth); err != nil {
			return err
		}

		return InvalidCodeErr
	}

	var addresses []*db.Address
	for _, v := range rq.Addresses {
		addresses = append(addresses, &db.Address{
			Id:      id,
			Address: v.Address,
			Network: v.Network,
		})
	}

	if err := c.db.CreateProfile(r.Context(), &db.Profile{
		Id:          id,
		PhoneNumber: rq.PhoneNumber,
	}, addresses); err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateProfile(r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := middleware.AuthId(r)

	profile, err := c.db.ProfileById(id)
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
		if !validators.IsValidName(rq.Name) {
			return InvalidPropertyErr
		}

		profile.Name = rq.Name
	}

	err = c.db.UpdateByPK(profile)
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
	if !validators.IsValidUsername(username) {
		return false, InvalidPropertyErr
	}

	isExist, err := c.db.UsernameIsExist(username)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

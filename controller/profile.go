package controller

import (
	"errors"
	"fmt"
	"fractapp-server/db"
	"fractapp-server/twilio"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

const (
	SMSMsg        = "Your code for Fractapp: %s"
	SMSTimeout    = 1 * time.Minute
	ResetCountSMS = 1 * time.Hour
	MaxSMSCount   = 3
)

var (
	InvalidSendSMSTimeoutErr = errors.New("expect timeout for new SMS sending")
)

type ProfileController struct {
	db     *pg.DB
	twilio *twilio.Api
}

func NewProfileController(db *pg.DB, fromNumber string, accountSid string, authToken string) *ProfileController {
	return &ProfileController{
		db:     db,
		twilio: twilio.NewApi(fromNumber, accountSid, authToken),
	}
}

func (controller *ProfileController) Registration(w http.ResponseWriter, r *http.Request) {
	if err := controller.registration(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Http error: %s \n", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (controller *ProfileController) registration(r *http.Request) error {
	phoneNumber := "+" + strings.Trim(r.URL.Query()["number"][0], " ")
	log.Printf("Phone number: %s\n", phoneNumber)

	if err := controller.twilio.ValidatePhoneNumber(phoneNumber); err != nil {
		return err
	}

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	codeInt := generator.Intn(999999)
	code := fmt.Sprintf("%d", codeInt)
	if len(code) < 6 {
		code = fmt.Sprintf("%06s", code)
	}

	auth := &db.Auth{}
	err := controller.db.Model(auth).Where("phone_number = ?", phoneNumber).Select()
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
		_, err = controller.db.Model(auth).Insert()
	} else {
		_, err = controller.db.Model(auth).
			Where("phone_number = ?", auth.PhoneNumber).
			Update()
	}

	if err := controller.twilio.SendSMS(phoneNumber, SMSMsg, code); err != nil {
		return err
	}

	return nil
}

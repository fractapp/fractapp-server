package notification

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	InvalidSendSMSErr     = errors.New("invalid send sms")
	InvalidPhoneNumberErr = errors.New("invalid phone number")
)

type Twilio struct {
	fromNumber string
	accountSid string
	authToken  string
}

func NewTwilioNotificator(fromNumber string, accountSid string, authToken string) Notificator {
	return &Twilio{
		fromNumber: fromNumber,
		accountSid: accountSid,
		authToken:  authToken,
	}
}

func (api *Twilio) Format(receiver string) string {
	return "+" + strings.ReplaceAll(receiver, " ", "")
}

func (api *Twilio) Validate(receiver string) error {
	urlStr := fmt.Sprintf("https://lookups.twilio.com/v1/PhoneNumbers/%s?Type=carrier", receiver)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(api.accountSid, api.authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Printf("SMS validate phone number: %s \n", bodyBytes)
		return InvalidPhoneNumberErr
	}

	return nil
}

func (api *Twilio) SendCode(receiver string, code string) error {
	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", api.accountSid)

	msgData := url.Values{}
	msgData.Set("From", api.fromNumber)
	msgData.Set("To", receiver)
	msgData.Set("Body", fmt.Sprintf("Your code for Fractapp: %s", code))
	msgDataReader := *strings.NewReader(msgData.Encode())

	req, err := http.NewRequest("POST", urlStr, &msgDataReader)
	if err != nil {
		return err
	}

	req.SetBasicAuth(api.accountSid, api.authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Printf("SMS service error: %s \n", bodyBytes)
		return InvalidSendSMSErr
	}

	return nil
}

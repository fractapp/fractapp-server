package twilio

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

type Api struct {
	fromNumber string
	accountSid string
	authToken  string
}

func NewApi(fromNumber string, accountSid string, authToken string) *Api {
	return &Api{
		fromNumber: fromNumber,
		accountSid: accountSid,
		authToken:  authToken,
	}
}

func (api *Api) SendSMS(phoneNumber string, msg string, code string) error {
	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", api.accountSid)

	msgData := url.Values{}
	msgData.Set("From", api.fromNumber)
	msgData.Set("To", phoneNumber)
	msgData.Set("Body", fmt.Sprintf(msg, code))
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

func (api *Api) ValidatePhoneNumber(phoneNumber string) error {
	urlStr := fmt.Sprintf("https://lookups.twilio.com/v1/PhoneNumbers/%s?Type=carrier", phoneNumber)

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

package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	TransactionApi     string
	SubstrateUrls      map[string]string
	BinanceApi         string
	SMSService         SMSService
	Firebase           Firebase
	DBConnectionString string
	Secret             string
	SMTP               `json:"SMTP"`
}

type SMTP struct {
	Host     string
	From     FromEmail
	Password string
}
type FromEmail struct {
	Name    string
	Address string
}

type SMSService struct {
	FromNumber string
	AccountSid string
	AuthToken  string
}
type Firebase struct {
	ProjectId string
}

func Parse(path string) (*Config, error) {
	config := &Config{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, config); err != nil {
		return nil, err
	}
	return config, nil
}

package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Origin        string
	SubstrateUrls map[string]string
	SMSService    SMSService
	Firebase      Firebase
	DB            DB
	Secret        string
	SMTP          `json:"SMTP"`
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
	ProjectId           string
	WithCredentialsFile string
}
type DB struct {
	Host     string
	User     string
	Password string
	Database string
}

func Parse(path string) (*Config, error) {
	config := &Config{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(file, config); err != nil {
		return config, err
	}
	return config, nil
}

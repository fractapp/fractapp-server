package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	SubstrateUrls       map[string]string
	ProjectId           string
	WithCredentialsFile string
	DB                  DB
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

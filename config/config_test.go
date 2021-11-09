package config

import (
	"testing"

	"gotest.tools/assert"
)

func TestParse(t *testing.T) {
	config, err := Parse("./test_files/config_test.json")
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, *config, Config{
		TransactionApi:     "txApi",
		BinanceApi:         "binanceApi",
		DBConnectionString: "dbConnection",
		SMSService: SMSService{
			FromNumber: "number",
			AccountSid: "sid",
			AuthToken:  "auth",
		},
		Firebase: Firebase{
			ProjectId: "projectId",
		},
		Secret: "secret",
		SMTP: SMTP{
			Host: "host",
			From: FromEmail{
				Name:    "name",
				Address: "address",
			},
			Password: "password",
		},
	})
}
func TestInvalidPath(t *testing.T) {
	_, err := Parse("./asdasdasd")
	assert.Error(t, err, "open ./asdasdasd: no such file or directory")
}

func TestInvalidJson(t *testing.T) {
	_, err := Parse("./test_files/invalid_config_test.json")
	assert.Error(t, err, "invalid character 'a' looking for beginning of value")
}

## Architecture



## Config
```
{
  "TransactionApi": "http://127.0.0.1:3000", // url from scanner api 
  "BinanceApi": "api.binance.com", // binance api url
  "Firebase": {
    "ProjectId": ""            // project id from firebase account
  },
  "SMSService": {
    "FromNumber": "",           // sender twilio number 
    "AccountSid": "",           // account sid from twilio account
    "AuthToken": ""             // aith token from twilio account
  },
  "DBConnectionString": {       // mongodb connection string
  "Secret": "",                 // secret for jwt token generator
  "SMTP": {                     // smtp server config 
    "Host": "",      
    "Password": "",             // smtp server password
    "From": { 
      "Name": "",               // Sender name 
      "Address": ""             // Sender email 
    }
  }
}
```

## Setup fractapp-server services

1. Setup firebase/postgres/twilio/smtp

2. [Install golang](https://golang.org/doc/install)

3. Copy firebase private key json to root project directory. This file must have name firebase.json. ([more info](https://firebase.google.com/docs/admin/setup))

4. Setup config.release.json (example in config.json)

## Run with golang

1. Build
```sh
make build 
```

2. Run api
```
./bin/api --host 0.0.0.0:9544 --config config.release.json

flags:
config - config file path
host - host for listen fractapp server
```

3. Run subscriber
```
./bin/subscriber --host 0.0.0.0:3005 --config config.release.json

flags:
config - config file path
host - host for listen fractapp server
```

4. Run scheduler
```
./bin/scheduler --config config.release.json

flags:
config - config file path
```

5. Run price saver for DOT
```
./bin/price --config config.release.json --currency DOT --start 1597622400000

flags:
config - config file path
currency - DOT/KSM
start - timestamp for start scan price
```

6. Run price saver for KSM
```
./bin/price --config config.release.json --currency KSM --start 1599177600000

flags:
config - config file path
currency - DOT/KSM
start - timestamp for start scan price
```

## Run with docker

Config for docker is in config-docker.json

1. Setup firebase/postgres/twilio/smtp
2. Copy firebase private key json to root project directory. This file must have name firebase.json. ([more info](https://firebase.google.com/docs/admin/setup))
3. Run docker-compose
```sh
docker-compose up
```

## Swagger

Swagger is available at {host}/swagger/index.html

## Make commands

Build
```sh
make build
```

Update all mocks in project for tests. (You need to have installed [mockgen](https://github.com/golang/mock))
```sh
make updateMocks
```

Start tests and get total coverage info
```sh
make totalCoverage
```

Start tests and get coverage in html format
```sh
make htmlCoverage
```
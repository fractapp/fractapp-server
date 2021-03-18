## Only android version

## Getting Started with golang

1. Install golang

   https://golang.org/doc/install

2. Setup firebase/postgres/twilio/smtp

3. Copy firebase private key json to root project directory. This file must have name firebase.json [more info](https://firebase.google.com/docs/admin/setup)
   
4. Setup config.release.json (example in config.json)
```
{
  "SubstrateUrls": {
    "Polkadot": "",              // wss host for Polkadot node (we can use "wss://rpc.polkadot.io")
    "Kusama": ""                 // wss host for Kusama node (we can use "wss://kusama-rpc.polkadot.io")
  },
  "Firebase": {
    "ProjectId": ""            // project id from firebase account
  },
  "SMSService": {
    "FromNumber": "",           // sender twilio number 
    "AccountSid": "",           // account sid from twilio account
    "AuthToken": ""             // aith token from twilio account
  },
  "DB": {                       // postgres config
    "Host": "",
    "User": "",
    "Password": "",
    "Database": ""
  },
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

5. Install packages
```sh
go mod download
```

6. Build
```sh
go build 
```

7. Run
```
./fractapp-server --config config.release.json --host 127.0.0.1:9544

flags:
config - config file path
host - host for listen fractapp server
```

## Docker

1. Setup firebase/postgres/twilio/smtp
   
2. Copy firebase private key json to root project directory. This file must have name firebase.json [more info](https://firebase.google.com/docs/admin/setup)

4. Setup config.release.json (example in config.json)
```
{
  "SubstrateUrls": {
    "Polkadot": "",              // wss host for Polkadot node (we can use "wss://rpc.polkadot.io")
    "Kusama": ""                 // wss host for Kusama node (we can use "wss://kusama-rpc.polkadot.io")
  },
  "Firebase": {
    "ProjectId": ""            // project id from firebase account
  },
  "SMSService": {
    "FromNumber": "",           // sender twilio number 
    "AccountSid": "",           // account sid from twilio account
    "AuthToken": ""             // aith token from twilio account
  },
  "DB": {                       // postgres config
    "Host": "",
    "User": "",
    "Password": "",
    "Database": ""
  },
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

5. Build
```
docker build -t fractapp-server .
```

6. Run
```
docker run -d -p 127.0.0.1:9544:9544 fractapp-server
```

## Swagger

Swagger is available at {host}/swagger/index.html

## Make commands

Update all mocks in project for tests
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

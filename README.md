## Only android version

## Getting Started with golang

1. Install golang

   https://golang.org/doc/install


2. Install packages
```sh
go mod download
```

3. Setup firebase/postgres/twilio/smtp
   
4. Setup config.json
```
{
  "SubstrateUrls": {
    "Polkadot": "",              // wss host for Polkadot node (we can use "wss://rpc.polkadot.io")
    "Kusama": ""                 // wss host for Kusama node (we can use "wss://kusama-rpc.polkadot.io")
  },
  "Firebase": {
    "ProjectId": "",            // project id from firebase account
    "WithCredentialsFile": ""   // Firebase SDK privKey [file](https://firebase.google.com/docs/admin/setup)  
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

4. Build
```sh
go build 
```

4. Run
```
./fractapp-server --config config.json --host 127.0.0.1:5000

flags:
config - config file path
host - host for listen fractapp server
```

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

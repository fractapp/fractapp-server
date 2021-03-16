## Authorization With Auth Public Key

Some methods use Auth Pub Key for authorization.
Auth Public Key is a subsidiary key from Main Private Key (Main Private Key -> Auth Private Key -> Auth Public Key).
It can be any keypair that a user has control. Now it is a subsidiary from the main seed.

For success, authorization needs to sign the message using the Auth Private Key.
Massage has this format: "It is my fractapp rq:{json request}{timestamp}"

Next, the signature/timestamp/authPubKey need to be added to headers:
Sign-Timestamp: timestamp that used in the signature
Sign: This signature
Auth-Key: Auth Public Key in hex format

## Sign in Fractapp

When a user wants authorization in fractapp then the user needs to use a code for confirmation that received on email or phone number.
User needs to create request to /auth/signIn:
```
{
    "Value": "",        // Email address or Phone number
    "Type": 0,          // Message type with code (0 - sms / 1 - email)
    "Addresses": {      // Addresses by network id (0 - polkadot/ 1 - kusama) from account
        0: {
            "Address": "",
            "PubKey": "",
            "Sign": ""
        }
    },    
    "Code": "000000"    // The code that was sent
}
```

All user addresses need to sign the message ("Sign" property in addresses):
"It is my auth key for fractapp:{Auth Public Key in hex format}{timestamp from auth with pub key}"
And these signatures need to put in the request.
Next, need to sign the request as described in the "Authorization With Auth Public Key" section.

## Subscribe

If a user wants to take notifications about transactions then the user needs to send the request to /notification/subscribe:
```
{
    "PubKey": "",
    "Address": "",
    "Network": 0,     // network id (0 - polkadot/ 1 - kusama)
    "Sign":"",
    "Token": "",      // firebase token
    "Timestamp": 0    // timestamp for signature
}
```

Sign property is signature for this message:
```
It is my firebase token for fractapp:{firebaseToken}{timestamp}
```

For this request not need to use the algorithm from the "Authorization With Auth Public Key" section.

## JWT Auth
JWT Auth use the header "Authorization" and has format:
```
BEARER {jwt token}
```

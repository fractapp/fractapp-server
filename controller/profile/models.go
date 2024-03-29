package profile

import (
	"fractapp-server/db"
	"fractapp-server/types"
)

type UpdateProfileRq struct {
	Name     string
	Username string
}
type MyProfile struct {
	Id          string `json:"id"`       // id from userInfo
	Name        string `json:"name"`     // name in fractapp
	Username    string `json:"username"` // username in fractapp
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	IsMigratory bool   `json:"isMigratory"` // always false. This property is for the future
	AvatarExt   string `json:"avatarExt"`   // avatar format (png/jpg/jpeg)
	LastUpdate  int64  `json:"lastUpdate"`  // timestamp of the last userInfo update
}
type ShortUserProfile struct {
	Id         string                   `json:"id"` // id from userInfo
	Name       string                   `json:"name"`
	Username   string                   `json:"username"`
	AvatarExt  string                   `json:"avatarExt"`  // avatar format (png/jpg/jpeg)
	IsChatBot  bool                     `json:"isChatBot"`  // always false. This property is for the future
	LastUpdate int64                    `json:"lastUpdate"` // timestamp of the last userInfo update
	Addresses  map[types.Network]string `json:"addresses"`  // String addresses by network (0 - polkadot/ 1 - kusama) from account
}

type TxStatusScannerApiRs struct {
	Status int64 `json:"status"`
}

type TxStatusRs struct {
	Hash   string `json:"hash"`
	Status int64  `json:"status"`
}

type MyContacts map[string]ShortUserProfile // map with id->short user userInfo

type UpdateFirebaseTokenRq struct {
	Token string `json:"token"`
}

type Transaction struct {
	ID        string         `json:"id"`
	Hash      string         `json:"hash"`
	Action    db.TxAction    `json:"action"`
	Currency  types.Currency `json:"currency"`
	To        string         `json:"to"`
	From      string         `json:"from"`
	Value     string         `json:"value"`
	Fee       string         `json:"fee"`
	Timestamp int64          `json:"timestamp"`
	Status    db.Status      `json:"status"`
}

type OldTransactionRs struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`

	Currency types.Currency `json:"currency"`

	From     string `json:"from"`
	UserFrom string `json:"userFrom"`

	Action db.TxAction `json:"action"`

	To     string `json:"to"`
	UserTo string `json:"userTo"`

	Value string  `json:"value"`
	Fee   string  `json:"fee"`
	Price float32 `json:"price"`

	Timestamp int64     `json:"timestamp"`
	Status    db.Status `json:"status"`
}

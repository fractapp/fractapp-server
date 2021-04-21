package profile

import "fractapp-server/types"

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
	Id         string                    `json:"id"` // id from userInfo
	Name       string                    `json:"name"`
	Username   string                    `json:"username"`
	AvatarExt  string                    `json:"avatarExt"`  // avatar format (png/jpg/jpeg)
	LastUpdate int64                     `json:"lastUpdate"` // timestamp of the last userInfo update
	Addresses  map[types.Currency]string `json:"addresses"`  // String addresses by network (0 - polkadot/ 1 - kusama) from account
}

type Transaction struct {
	ID       string `json:"id"`
	Currency int    `json:"currency"`

	From     string `json:"from"`
	UserFrom string `json:"userFrom"`

	To     string `json:"to"`
	UserTo string `json:"userTo"`

	Value      string  `json:"value"`
	UsdValue   float32 `json:"usdValue"`
	FloatValue string  `json:"floatValue"`

	Fee      string  `json:"fee"`
	UsdFee   float32 `json:"usdFee"`
	FloatFee string  `json:"floatFee"`

	Timestamp int64 `json:"timestamp"`
	Status    int64 `json:"status"`
}

type TxStatusRs struct {
	Status int64 `json:"status"`
}

type Balance struct {
	Value string `json:"value"`
}

type MyContacts map[string]ShortUserProfile // map with id->short user userInfo

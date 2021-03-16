package profile

import "fractapp-server/types"

type UpdateProfileRq struct {
	Name     string
	Username string
}
type MyProfile struct {
	Id          string `json:"id"`       // id from info
	Name        string `json:"name"`     // name in fractapp
	Username    string `json:"username"` // username in fractapp
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"notification"`
	IsMigratory bool   `json:"isMigratory"` // always false. This property is for the future
	AvatarExt   string `json:"avatarExt"`   // avatar format (png/jpg/jpeg)
	LastUpdate  int64  `json:"lastUpdate"`  // timestamp of the last info update
}
type ShortUserProfile struct {
	Id         string                    `json:"id"` // id from info
	Name       string                    `json:"name"`
	Username   string                    `json:"username"`
	AvatarExt  string                    `json:"avatarExt"`  // avatar format (png/jpg/jpeg)
	LastUpdate int64                     `json:"lastUpdate"` // timestamp of the last info update
	Addresses  map[types.Currency]string `json:"addresses"`  // String addresses by network (0 - polkadot/ 1 - kusama) from account
}

type MyContacts map[string]ShortUserProfile // map with id->short user info

package profile

import "fractapp-server/types"

type UpdateProfileRq struct {
	Name     string
	Username string
}
type MyProfile struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"notification"`
	IsMigratory bool   `json:"isMigratory"`
	AvatarExt   string `json:"avatarExt"`
	LastUpdate  int64  `json:"lastUpdate"`
}
type UserProfileShort struct {
	Id         string                    `json:"id"`
	Name       string                    `json:"name"`
	Username   string                    `json:"username"`
	AvatarExt  string                    `json:"avatarExt"`
	LastUpdate int64                     `json:"lastUpdate"`
	Addresses  map[types.Currency]string `json:"addresses"`
}

type MyContacts map[string]UserProfileShort

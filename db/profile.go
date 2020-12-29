package db

type Profile struct {
	Id          string
	Name        string
	Username    string
	PhoneNumber string `pg:"phone_number"`
	Email       string
	Twitter     string
}

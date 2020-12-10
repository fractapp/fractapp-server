package db

type Auth struct {
	PhoneNumber string `pg:"phone_number"`
	Code        string
	Count       int64 `pg:",use_zero"`
	Timestamp   int64
}

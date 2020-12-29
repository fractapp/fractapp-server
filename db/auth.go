package db

type Auth struct {
	PhoneNumber string `pg:"phone_number"`
	Code        string
	Attempts    int32
	Count       int32 `pg:",use_zero"`
	Timestamp   int64
}

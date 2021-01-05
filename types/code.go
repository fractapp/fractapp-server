package types

type CodeType int

const (
	PhoneNumberCode CodeType = iota
	EmailCode
	TwitterCode
)

type CheckType int

const (
	Auth CheckType = iota
	Migration
)

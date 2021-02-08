package types

type CodeType int

const (
	PhoneNumberCode CodeType = iota
	EmailCode
)

type CheckType int

const (
	Auth CheckType = iota
	Change
)

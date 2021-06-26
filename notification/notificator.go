package notification

type NotificatorType int
type CheckType int

const (
	SMS NotificatorType = iota
	Email
	CryptoAddress
)
const (
	Auth CheckType = iota
)

type Notificator interface {
	Format(receiver string) string
	Validate(receiver string) error
	SendCode(receiver string, code string) error
}

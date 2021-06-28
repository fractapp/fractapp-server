package notification

type NotificatorType int

const (
	SMS NotificatorType = iota
	Email
	CryptoAddress
)

type Notificator interface {
	Format(receiver string) string
	Validate(receiver string) error
	SendCode(receiver string, code string) error
}

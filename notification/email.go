package notification

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

var (
	InvalidSendEmailErr = errors.New("invalid send notification")
	InvalidEmailErr     = errors.New("invalid email address")
)

const (
	MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type SMTP struct {
	auth smtp.Auth
	host string
	from mail.Address
}

func NewSMTPNotificator(host string, address string, name string, password string) (Notificator, error) {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	auth := smtp.PlainAuth("", address, password, hostOnly)
	return &SMTP{
		auth: auth,
		from: mail.Address{Name: name, Address: address},
		host: host,
	}, nil
}

func (client *SMTP) Format(receiver string) string {
	receiver = strings.ReplaceAll(receiver, " ", "")
	return strings.ToLower(receiver)
}

func (client *SMTP) Validate(receiver string) error {
	if strings.Count(receiver, "@") == 0 {
		return InvalidEmailErr
	}
	return nil
}

func (client *SMTP) SendCode(receiver string, code string) error {
	subj := "Activation code: " + code

	t, err := template.ParseFiles("templates/auth.html")
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, map[string]string{"code": code}); err != nil {
		return err
	}
	body := buffer.String()

	err = client.send(receiver, subj, body)
	if err != nil {
		return err
	}

	return nil
}

func (client *SMTP) send(to string, subj string, body string) error {
	headers := make(map[string]string)
	headers["From"] = client.from.String()
	headers["To"] = to
	headers["Subject"] = subj

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += MIME + "\r\n" + body

	err := smtp.SendMail(client.host, client.auth, client.from.Address, []string{to}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

package email

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
	InvalidSendEmailErr = errors.New("invalid send email")
	InvalidEmailErr     = errors.New("invalid email address")
)

const (
	MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type Client struct {
	auth smtp.Auth
	host string
	from mail.Address
}

func New(host string, address string, name string, password string) (*Client, error) {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	auth := smtp.PlainAuth("", address, password, hostOnly)
	return &Client{
		auth: auth,
		from: mail.Address{Name: name, Address: address},
		host: host,
	}, nil
}

func (client *Client) FormatEmail(value string) string {
	value = strings.ReplaceAll(value, " ", "")
	return strings.ToLower(value)
}

func (client *Client) ValidateEmail(value string) error {
	if strings.Count(value, "@") == 0 {
		return InvalidEmailErr
	}
	return nil
}

func (client *Client) SendCode(to string, code string) error {
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

	err = client.send(to, subj, body)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) send(to string, subj string, body string) error {
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

package firebase

import (
	"context"
	"fmt"
	"fractapp-server/types"
	"log"

	"google.golang.org/api/option"

	"firebase.google.com/go/messaging"

	firebase "firebase.google.com/go"
)

type TxType int

const (
	Sent TxType = iota
	Received
)

type TxNotificator interface {
	Notify(msg string, token string) error
	Msg(member string, txType TxType, amount float64, currency types.Currency) string
}

type Client struct {
	ctx       context.Context
	msgClient *messaging.Client
}

func NewClient(ctx context.Context, credentialsFile string, projectId string) (*Client, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	config := &firebase.Config{ProjectID: projectId}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	msg, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:       ctx,
		msgClient: msg,
	}, nil
}

func (n *Client) Notify(msg string, token string) error {
	n.msgClient.Send(n.ctx, &messaging.Message{
		Notification: &messaging.Notification{
			Body: msg,
		},
		Token: token,
	})
	return nil
}

func (n *Client) Msg(member string, txType TxType, amount float64, currency types.Currency) string {
	switch txType {
	case Sent:
		return fmt.Sprintf("You sent %.3f %s to %s", amount, currency.String(), member)
	case Received:
		return fmt.Sprintf("You received %.3f %s from %s", amount, currency.String(), member)
	}

	return ""
}

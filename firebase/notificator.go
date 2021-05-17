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
	Notify(title string, msg string, token string) error
	MsgForAuthed(txType TxType, amount float64, currency types.Currency) string
	MsgForAddress(address string, txType TxType, amount float64, currency types.Currency) string
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

func (n *Client) Notify(title string, msg string, token string) error {
	_, err := n.msgClient.Send(n.ctx, &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  msg,
		},
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				ChannelID: "chats",
				Priority:  messaging.PriorityDefault,
			},
		},
		Token: token,
	})
	if err != nil {
		return err
	}
	return nil
}

func (n *Client) MsgForAuthed(txType TxType, amount float64, currency types.Currency) string {
	switch txType {
	case Sent:
		return fmt.Sprintf("You sent %.3f %s", amount, currency.String())
	case Received:
		return fmt.Sprintf("You received %.3f %s", amount, currency.String())
	}

	return ""
}

func (n *Client) MsgForAddress(address string, txType TxType, amount float64, currency types.Currency) string {
	switch txType {
	case Sent:
		return fmt.Sprintf("You sent %.3f %s to %s", amount, currency.String(), address)
	case Received:
		return fmt.Sprintf("You received %.3f %s from %s", amount, currency.String(), address)
	}

	return ""
}

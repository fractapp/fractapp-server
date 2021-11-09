package push

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

type Notificator interface {
	Notify(title string, msg string, token string) error
	CreateMsg(txType TxType, amount float64, usdAmount float64, currency types.Currency) string
}

type Client struct {
	ctx       context.Context
	msgClient *messaging.Client
}

func NewClient(ctx context.Context, credentialsFile string, projectId string) (Notificator, error) {
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
func (n *Client) CreateMsg(txType TxType, amount float64, usdAmount float64, currency types.Currency) string {
	amountMsg := fmt.Sprintf("%.2f (%.4f %s)", usdAmount, amount, currency.String())
	if usdAmount/100 < 1 {
		amountMsg = fmt.Sprintf("%.4f %s", amount, currency.String())
	}
	switch txType {
	case Sent:
		return fmt.Sprintf("You sent $%s", amountMsg)
	case Received:
		return fmt.Sprintf("You received $%s", amountMsg)
	}

	return ""
}

package message

import (
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	"fractapp-server/types"
)

type Action string

type MessageRq struct {
	Value    string            `json:"value"`
	Action   string            `json:"action"`
	Receiver string            `json:"receiver"`
	Args     map[string]string `json:"args"`
	Rows     []db.Row          `json:"rows"`
}

type TransactionRs struct {
	Id            string         `json:"id"`
	Hash          string         `json:"hash"`
	Currency      types.Currency `json:"currency"`
	MemberAddress string         `json:"memberAddress"`
	Member        *string        `json:"member"`
	Direction     db.TxDirection `json:"direction"`
	Action        db.TxAction    `json:"action"`
	Status        db.Status      `json:"status"`
	Value         string         `json:"value"`
	Fee           string         `json:"fee"`
	Price         float32        `json:"price"`
	Timestamp     int64          `json:"timestamp"`
}

type MessagesAndTxs struct {
	Messages    []MessageRs                         `json:"messages"`
	Transaction []db.Transaction                    `json:"transactions"`
	Users       map[string]profile.ShortUserProfile `json:"users"`
}

type SendInfo struct {
	Timestamp int64 `json:"timestamp"`
}

type MessageRs struct {
	Id string `json:"id"`

	Version int               `json:"version"`
	Value   string            `json:"value"`
	Action  Action            `json:"action"`
	Args    map[string]string `json:"args"`
	Rows    []db.Row          `json:"rows"`

	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Timestamp int64  `json:"timestamp"`
}

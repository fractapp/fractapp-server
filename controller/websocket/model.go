package websocket

import (
	"fractapp-server/controller/info"
	"fractapp-server/controller/message"
	"fractapp-server/controller/profile"
	"fractapp-server/controller/substrate"
	"fractapp-server/types"
)

type Method string
type RsMethod string

const (
	setDeliveredMethod   Method = "set_delivered"
	getUsersMethod       Method = "get_users"
	getTxsStatusesMethod Method = "get_txs_statuses"

	updateMethod      RsMethod = "update"
	balancesMethod    RsMethod = "balances"
	txsStatusesMethod RsMethod = "txs_statuses"
	usersMethod       RsMethod = "users"
)

type Rq struct {
	Method Method   `json:"method"`
	Ids    []string `json:"ids"`
}

type WsResponse struct {
	Method RsMethod    `json:"method"`
	Value  interface{} `json:"value"`
}

type Update struct {
	Transactions  map[types.Currency][]*message.TransactionRs `json:"transactions"`
	Messages      []*message.MessageRs                        `json:"messages"`
	Users         map[string]profile.ShortUserProfile         `json:"users"`
	Notifications []string                                    `json:"notifications"`
	Prices        []*info.Price                               `json:"prices"`
}

type Balances struct {
	Balances map[types.Currency]*substrate.Balance `json:"balances"`
}

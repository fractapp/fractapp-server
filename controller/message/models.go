package message

import (
	"fractapp-server/controller/profile"
	"fractapp-server/db"
)

const (
	Init Action = "init"
)

type Action string

type MessageRq struct {
	Version  int      `json:"version"`
	Value    string   `json:"value"`
	Action   string   `json:"action"`
	Receiver string   `json:"receiver"`
	Args     []string `json:"args"`
	Rows     []db.Row `json:"rows"`
}

type MessagesAndUsersRs struct {
	Messages []MessageRs                         `json:"messages"`
	Users    map[string]profile.ShortUserProfile `json:"users"`
}

type MessageRs struct {
	Id string `json:"id"`

	Version int      `json:"version"`
	Value   string   `json:"value"`
	Action  Action   `json:"action"`
	Args    []string `json:"args"`
	Rows    []db.Row `json:"rows"`

	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Timestamp int64  `json:"timestamp"`
}

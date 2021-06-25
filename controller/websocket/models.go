package websocket

import "fractapp-server/controller/profile"

const (
	Messages Method = "messages"
	Send     Method = "send"
)

type Method string

type Request struct {
	Id        int64      `json:"id"`
	Method    Method     `json:"method"`
	Message   *MessageRq `json:"message"`
	Delivered *[]string  `json:"changeMsgStatus"`
}

type Response struct {
	Id     int64       `json:"id"`
	Ok     bool        `json:"ok"`
	Method Method      `json:"method"`
	Value  interface{} `json:"value"`
}

type Error struct {
	Error string `json:"error"`
}

type MessageRq struct {
	Version  int    `json:"version"`
	Value    string `json:"value"`
	Rows     []Row  `json:"rows"`
	Receiver string `json:"receiver"`
}

type MessagesAndUsersRs struct {
	Messages []MessageRs                         `json:"messages"`
	Users    map[string]profile.ShortUserProfile `json:"users"`
}
type MessageAndUserRs struct {
	Message MessageRs                `json:"message"`
	User    profile.ShortUserProfile `json:"user"`
}
type MessageRs struct {
	Id string `json:"id"`

	Version int    `json:"version"`
	Value   string `json:"value"`
	Rows    []Row  `json:"rows"`

	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
}

type Row struct {
	Buttons []ButtonRq
}

type ButtonRq struct {
	Value     string   `json:"value"`
	Action    string   `json:"action"`
	Arguments []string `json:"arguments"`
	ImageUrl  string   `json:"image_url"`
}

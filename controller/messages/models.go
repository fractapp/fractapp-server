package messages

type Message struct {
	Id        string   `json:"id"`
	Value     string   `json:"value"`
	Buttons   []Button `json:"buttons"`
	SenderId  string   `json:"senderId"`
	Timestamp int64    `json:"timestamp"`
}

type Button struct {
	Id     string `json:"id"`
	Value  string `json:"value"`
	Action string `json:"action"`
	Img    string `json:"img"`
}

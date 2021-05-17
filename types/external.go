package types

type Transaction struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	Currency  int    `json:"currency"`
	To        string `json:"to"`
	From      string `json:"from"`
	Value     string `json:"value"`
	Fee       string `json:"fee"`
	Timestamp int64  `json:"timestamp"`
	Status    int64  `json:"status"`
}

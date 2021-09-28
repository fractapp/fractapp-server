package substrate

type FeeInfo struct {
	Fee string `json:"fee"`
}

type BroadcastResult struct {
	Hash string `json:"hash"`
}

type Balance struct {
	Total         string `json:"total"`
	Transferable  string `json:"transferable"`
	PayableForFee string `json:"payableForFee"`
	Staking       string `json:"staking"`
}

type TxBase struct {
	BlockNumber int64  `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	Nonce       int64  `json:"nonce"`
}

type Base struct {
	GenesisHash        string `json:"genesisHash"`
	Metadata           string `json:"metadata"`
	SpecVersion        int64  `json:"specVersion"`
	TransactionVersion int64  `json:"transactionVersion"`
}

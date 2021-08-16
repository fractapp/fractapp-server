package substrate

type FeeInfo struct {
	Fee string `json:"fee"`
}

type BroadcastResult struct {
	Hash string `json:"hash"`
}

type Balance struct {
	Value string `json:"value"`
}

type TxBase struct {
	BlockNumber int64  `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	GenesisHash string `json:"genesisHash"`
	Metadata    string `json:"metadata"`

	SpecVersion        int64 `json:"specVersion"`
	TransactionVersion int64 `json:"transactionVersion"`
	Nonce              int64 `json:"nonce"`
}

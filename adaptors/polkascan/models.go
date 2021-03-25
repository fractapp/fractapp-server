package polkascan

type Block struct {
	Errors []interface{} `json:"errors"`
	Data   struct {
		Type       string `json:"type"`
		ID         int    `json:"id"`
		Attributes struct {
			ID   int    `json:"id"`
			Hash string `json:"hash"`
		} `json:"attributes"`
	} `json:"data"`
	Included []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			BlockID       int    `json:"block_id"`
			EventIdx      int    `json:"event_idx"`
			ExtrinsicIdx  int    `json:"extrinsic_idx"`
			Type          string `json:"type"`
			SpecVersionID int    `json:"spec_version_id"`
			ModuleID      string `json:"module_id"`
			EventID       string `json:"event_id"`
			System        int    `json:"system"`
			Module        int    `json:"module"`
			Phase         int    `json:"phase"`
			Attributes    []struct {
				Type     string      `json:"type"`
				Value    interface{} `json:"value"`
				Valueraw string      `json:"valueRaw"`
			} `json:"attributes"`
			CodecError bool `json:"codec_error"`
		} `json:"attributes"`
	} `json:"included"`
}

type LatestInfo struct {
	Meta struct {
		Authors []string `json:"authors"`
	} `json:"meta"`
	Errors []interface{} `json:"errors"`
	Data   struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			BestBlock             int    `json:"best_block"`
			TotalSignedExtrinsics int    `json:"total_signed_extrinsics"`
			TotalEvents           int    `json:"total_events"`
			TotalEventsModule     int    `json:"total_events_module"`
			TotalBlocks           string `json:"total_blocks"`
			TotalAccounts         int    `json:"total_accounts"`
			TotalRuntimes         int    `json:"total_runtimes"`
		} `json:"attributes"`
	} `json:"data"`
	Links struct {
	} `json:"links"`
}

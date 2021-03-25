package polkascan

import (
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/types"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/centrifuge/go-substrate-rpc-client/v2/rpc/chain"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
)

type Adaptor struct {
	host       string
	api        *gsrpc.SubstrateAPI
	blockEvent *chain.FinalizedHeadsSubscription
	network    types.Network
}

func NewAdaptor(host string, network types.Network) *Adaptor {
	return &Adaptor{host: host, network: network}
}

func (a *Adaptor) FetchTimeoutSec() int64 {
	return 1
}

func (a *Adaptor) LastHeight() (uint64, error) {
	r, err := http.Get(fmt.Sprintf("%s/networkstats/latest", a.host))
	if err != nil {
		return 0, err
	}
	if r.StatusCode != http.StatusOK {
		return 0, errors.New("status code not equals OK (200)")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}

	info := LatestInfo{}
	err = json.Unmarshal(b, &info)
	if err != nil {
		return 0, err
	}

	return uint64(info.Data.Attributes.BestBlock), nil
}

func (a *Adaptor) Transfers(blockNumber uint64) ([]types.Tx, error) {
	r, err := http.Get(fmt.Sprintf("%s/block/%d?include=events", a.host, blockNumber))
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, errors.New("status code not equals OK (200)")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	block := Block{}
	err = json.Unmarshal(b, &block)
	if err != nil {
		return nil, err
	}

	var txs []types.Tx
	for _, v := range block.Included {
		if v.Type != "event" || v.Attributes.ModuleID != "balances" || v.Attributes.EventID != "Transfer" {
			continue
		}
		if len(v.Attributes.Attributes) < 3 {
			continue
		}

		sender := v.Attributes.Attributes[0]
		receiver := v.Attributes.Attributes[1]
		value := v.Attributes.Attributes[2]

		if sender.Type != "AccountId" || receiver.Type != "AccountId" {
			continue
		}
		txs = append(txs, types.Tx{
			Sender:     sender.Value.(string),
			Receiver:   receiver.Value.(string),
			FullAmount: big.NewInt(int64(value.Value.(float64))),
		})
	}

	return nil, nil
}

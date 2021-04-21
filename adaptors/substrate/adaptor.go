package substrate

import (
	"fmt"
	ftypes "fractapp-server/types"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/v2/rpc/chain"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
)

type Adaptor struct {
	host       string
	api        *gsrpc.SubstrateAPI
	blockEvent *chain.FinalizedHeadsSubscription
	network    ftypes.Network
}

func NewAdaptor(host string, network ftypes.Network) *Adaptor {
	return &Adaptor{host: host, network: network}
}

func (a *Adaptor) Connect() error {
	api, err := gsrpc.NewSubstrateAPI(a.host)
	if err != nil {
		return err
	}
	a.api = api

	return nil
}

func (a *Adaptor) Subscribe() error {
	newBlockEvent, err := a.api.RPC.Chain.SubscribeFinalizedHeads()
	if err != nil {
		return err
	}
	a.blockEvent = newBlockEvent

	return nil
}

func (a *Adaptor) Unsubscribe() {
	a.blockEvent.Unsubscribe()
}

func (a *Adaptor) WaitNewBlock() (uint64, error) {
	select {
	case e := <-a.blockEvent.Chan():
		return uint64(e.Number), nil
	case err := <-a.blockEvent.Err():
		return 0, err
	}
}

func (a *Adaptor) Err() <-chan error {
	return a.blockEvent.Err()
}

func (a *Adaptor) Transfers(blockNumber uint64) ([]ftypes.Tx, error) {
	hash, err := a.api.RPC.Chain.GetBlockHash(blockNumber)
	if err != nil {
		return nil, err
	}
	meta, err := a.api.RPC.State.GetMetadata(hash)
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return nil, err
	}

	raw, err := a.api.RPC.State.GetStorageRaw(key, hash)
	if err != nil {
		return nil, err
	}

	events := types.EventRecords{}
	err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	if err != nil {
		return nil, err
	}

	fees := make(map[uint32]*big.Int)
	for _, v := range events.Treasury_Deposit {
		fees[v.Phase.AsApplyExtrinsic] = v.Deposited.Int
	}
	for _, v := range events.Balances_Deposit {
		if fee, ok := fees[v.Phase.AsApplyExtrinsic]; ok {
			fees[v.Phase.AsApplyExtrinsic] = fee.Add(fee, v.Balance.Int)
		} else {
			fees[v.Phase.AsApplyExtrinsic] = v.Balance.Int
		}
	}

	var txs []ftypes.Tx
	for _, v := range events.Balances_Transfer {
		txs = append(txs, ftypes.Tx{
			EventID:    fmt.Sprintf("%d-%d", blockNumber, v.Phase.AsApplyExtrinsic),
			Sender:     a.network.Address(v.From[:]),
			Receiver:   a.network.Address(v.To[:]),
			FullAmount: v.Value.Int,
			Fee:        fees[v.Phase.AsApplyExtrinsic],
		})
	}

	return txs, nil
}

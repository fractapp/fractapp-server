package adaptors

import (
	ftypes "fractapp-server/types"

	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	log "github.com/sirupsen/logrus"

	"github.com/centrifuge/go-substrate-rpc-client/v2/rpc/chain"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
)

type SubstrateAdaptor struct {
	host       string
	api        *gsrpc.SubstrateAPI
	blockEvent *chain.FinalizedHeadsSubscription
	network    ftypes.Network
}

func NewSubstrateAdaptor(host string, network ftypes.Network) *SubstrateAdaptor {
	return &SubstrateAdaptor{host: host, network: network}
}

func (a *SubstrateAdaptor) Connect() error {
	api, err := gsrpc.NewSubstrateAPI(a.host)
	if err != nil {
		return err
	}
	a.api = api

	return nil
}

func (a *SubstrateAdaptor) Subscribe() error {
	newBlockEvent, err := a.api.RPC.Chain.SubscribeFinalizedHeads()
	if err != nil {
		return err
	}
	a.blockEvent = newBlockEvent

	return nil
}

func (a *SubstrateAdaptor) Unsubscribe() {
	a.blockEvent.Unsubscribe()
}

func (a *SubstrateAdaptor) LastHeight() (uint64, error) {
	select {
	case e := <-a.blockEvent.Chan():
		return uint64(e.Number), nil
	case err := <-a.blockEvent.Err():
		return 0, err
	}
}

func (a *SubstrateAdaptor) Err() <-chan error {
	return a.blockEvent.Err()
}

func (a *SubstrateAdaptor) Transfers(blockNumber uint64) ([]ftypes.Tx, error) {
	hash, err := a.api.RPC.Chain.GetBlockHash(blockNumber)
	if err != nil {
		return nil, err
	}
	a.api.RPC.Chain.GetBlockHashLatest()
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

	//fees := make(map[types.Hash]*big.Int)
	for _, v := range events.Treasury_Deposit {
		log.Info(len(v.Topics))
		for _, c := range v.Topics {
			log.Info(c.Hex())
		}
		//fees[v.Topics] = v.Deposited.Int
	}
	for _, v := range events.Balances_Deposit {
		log.Info(len(v.Topics))
		for _, c := range v.Topics {
			log.Info(c.Hex())
		}
		//fees[v.Topics] = v.Deposited.Int
	}
	var txs []ftypes.Tx
	for _, v := range events.Balances_Transfer {
		log.Info(len(v.Topics))
		for _, c := range v.Topics {
			log.Info(c.Hex())
		}

		txs = append(txs, ftypes.Tx{
			Sender:     a.network.Address(v.From[:]),
			Receiver:   a.network.Address(v.To[:]),
			FullAmount: v.Value.Int,
		})
	}

	return txs, nil
}

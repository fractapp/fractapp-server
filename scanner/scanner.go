package scanner

import (
	"fractapp-server/db"
	"fractapp-server/notificator"
	"log"
	"math/big"

	dbType "fractapp-server/types"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

const addressesLimit = 100000

type EventScanner struct {
	host        string
	db          db.DB
	prefix      string
	network     dbType.Network
	notificator notificator.Notificator
}

func NewEventScanner(host string, db db.DB, prefix string, network dbType.Network, notificator notificator.Notificator) *EventScanner {
	return &EventScanner{
		host:        host,
		db:          db,
		prefix:      prefix,
		network:     network,
		notificator: notificator,
	}
}

func (s *EventScanner) Start() error {
	api, err := gsrpc.NewSubstrateAPI(s.host)
	if err != nil {
		return err
	}

	log.Printf("%s: subscribe new block \n", s.prefix)
	newBlockEvent, err := api.RPC.Chain.SubscribeFinalizedHeads()
	if err != nil {
		return err
	}

	defer newBlockEvent.Unsubscribe()

	var lastHeight uint64
	for {
		select {
		case newHeader := <-newBlockEvent.Chan():
			blockNumber := uint64(newHeader.Number)
			if lastHeight > blockNumber {
				continue
			}
			err := s.scanBlock(api, blockNumber)
			if err != nil {
				log.Printf("%s: Error scan block: %s \n", s.prefix, err.Error())
			}
			lastHeight = blockNumber
		case err := <-newBlockEvent.Err():
			log.Printf("%s: Error substrate rpc: %s \n", s.prefix, err.Error())
			log.Printf("%s: Repeated subscribe new block \n", s.prefix)

			newBlockEvent.Unsubscribe()
			newBlockEvent, err = api.RPC.Chain.SubscribeFinalizedHeads()
			if err != nil {
				log.Printf("%s: Error repeated subscribe: %s \n", s.prefix, err.Error())
			}
			continue
		}
	}
}

func (s *EventScanner) scanBlock(api *gsrpc.SubstrateAPI, number uint64) error {
	log.Printf("%s: Scan new block: %d \n", s.prefix, number)

	hash, err := api.RPC.Chain.GetBlockHash(number)
	if err != nil {
		return err
	}

	meta, err := api.RPC.State.GetMetadata(hash)
	if err != nil {
		return err
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return err
	}

	raw, err := api.RPC.State.GetStorageRaw(key, hash)
	if err != nil {
		return err
	}

	events := types.EventRecords{}
	err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	if err != nil {
		return err
	}

	senders := make(map[string]map[string]*big.Int)
	receivers := make(map[string]map[string]*big.Int)
	for _, v := range events.Balances_Transfer {
		sender := s.network.Address(v.From[:])
		receiver := s.network.Address(v.To[:])
		if _, ok := senders[sender]; !ok {
			senders[sender] = make(map[string]*big.Int)
		}
		if _, ok := receivers[receiver]; !ok {
			receivers[receiver] = make(map[string]*big.Int)
		}

		if _, ok := senders[sender][receiver]; !ok {
			senders[sender][receiver] = big.NewInt(0)
		}
		if _, ok := receivers[receiver][sender]; !ok {
			receivers[receiver][sender] = big.NewInt(0)
		}

		senders[sender][receiver].Add(senders[sender][receiver], v.Value.Int)
		receivers[receiver][sender].Add(receivers[receiver][sender], v.Value.Int)
	}

	addrCount, err := s.db.SubscribersCount()
	if err != nil {
		return err
	}

	if len(receivers) == 0 || len(senders) == 0 {
		return nil
	}

	i := 0
	for i < addrCount {
		subscribers, err := s.db.SubscribersByRange(i, addressesLimit)
		if err != nil {
			return err
		}

		i += addressesLimit

		for _, sub := range subscribers {
			if len(receivers) == 0 || len(senders) == 0 {
				return nil
			}

			sentTxs, senderExist := senders[sub.Address]
			if senderExist {
				for receiver, amount := range sentTxs {
					currency := s.network.Currency()

					name := receiver
					p, err := s.db.ProfileByAddress(receiver)
					if err != nil && err != db.ErrNoRows {
						log.Printf("invalid get by address in notification service: %s\n", err.Error())
						continue
					}
					if err != db.ErrNoRows {
						name = p.Name
					}

					msg := s.notificator.Msg(name, notificator.Sent, currency.ConvertFromPlanck(amount), currency)
					err = s.notificator.Notify(msg, sub.Token)

					log.Printf("%s: Notify Type: Sent; Sender:%s; Receiver:%s; Sub:%s Amount:%s; \n",
						s.prefix, sub.Address, receiver, sub.Address, amount.String())
					if err != nil {
						log.Printf("%s: Error: %s \n", s.prefix, err.Error())
					}
				}
			}
			delete(senders, sub.Address)

			receivedTxs, receiverExist := receivers[sub.Address]
			if receiverExist {
				for sender, amount := range receivedTxs {
					currency := s.network.Currency()

					name := sender
					p, err := s.db.ProfileByAddress(sender)
					if err != nil && err != db.ErrNoRows {
						log.Printf("invalid get by address in notification service: %s\n", err.Error())
						continue
					}
					if err != db.ErrNoRows {
						name = p.Name
					}

					msg := s.notificator.Msg(name, notificator.Received, currency.ConvertFromPlanck(amount), currency)
					err = s.notificator.Notify(msg, sub.Token)

					log.Printf("%s: Notify Type: Received; Sender:%s; Receiver:%s; Sub:%s Amount:%s; \n",
						s.prefix, sender, sub.Address, sub.Address, amount.String())
					if err != nil {
						log.Printf("%s: Error: %s \n", s.prefix, err.Error())
					}
				}
			}
			delete(receivers, sub.Address)
		}
	}

	return nil
}

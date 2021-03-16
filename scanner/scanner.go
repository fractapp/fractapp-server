package scanner

import (
	"fractapp-server/adaptors"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"log"
	"math/big"

	dbType "fractapp-server/types"
)

const addressesLimit = 100000

type BlockScanner struct {
	db          db.DB
	prefix      string
	network     dbType.Network
	notificator firebase.TxNotificator
	adaptor     adaptors.Adaptor
}

func NewBlockScanner(db db.DB, prefix string, network dbType.Network, notificator firebase.TxNotificator, adaptor adaptors.Adaptor) *BlockScanner {
	return &BlockScanner{
		db:          db,
		prefix:      prefix,
		network:     network,
		notificator: notificator,
		adaptor:     adaptor,
	}
}

func (s *BlockScanner) Start() error {
	err := s.adaptor.Connect()
	if err != nil {
		return err
	}

	log.Printf("%s: subscribe new block \n", s.prefix)
	err = s.adaptor.Subscribe()
	if err != nil {
		return err
	}

	defer s.adaptor.Unsubscribe()

	var lastHeight uint64
	for {
		lastHeight = s.scanNewHeight(lastHeight)
	}
}

func (s *BlockScanner) scanNewHeight(lastHeight uint64) uint64 {
	blockNumber, err := s.adaptor.WaitNewBlock()
	if err != nil {
		log.Printf("%s: Error substrate rpc: %s \n", s.prefix, err.Error())
		log.Printf("%s: Repeated subscribe new block \n", s.prefix)

		s.adaptor.Unsubscribe()
		err = s.adaptor.Subscribe()

		if err != nil {
			log.Printf("%s: Error repeated subscribe: %s \n", s.prefix, err.Error())
		}
		return lastHeight
	} else {
		if lastHeight > blockNumber {
			return lastHeight
		}
		err := s.scanBlock(blockNumber)
		if err != nil {
			log.Printf("%s: Error scan block: %s \n", s.prefix, err.Error())
		}
		return blockNumber
	}
}

func (s *BlockScanner) scanBlock(number uint64) error {
	log.Printf("%s: Scan new block: %d \n", s.prefix, number)

	transfers, err := s.adaptor.Transfers(number)
	if err != nil {
		return err
	}

	senders := make(map[string]map[string]*big.Int)
	receivers := make(map[string]map[string]*big.Int)
	for _, v := range transfers {
		sender := v.Sender
		receiver := v.Receiver
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

		senders[sender][receiver].Add(senders[sender][receiver], v.FullAmount)
		receivers[receiver][sender].Add(receivers[receiver][sender], v.FullAmount)
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
			if len(receivers) == 0 && len(senders) == 0 {
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

					msg := s.notificator.Msg(name, firebase.Sent, currency.ConvertFromPlanck(amount), currency)
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

					msg := s.notificator.Msg(name, firebase.Received, currency.ConvertFromPlanck(amount), currency)
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

package scanner

import (
	"fractapp-server/adaptors"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"math/big"

	log "github.com/sirupsen/logrus"

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
	log.Infof("%s: start scanner ...", s.prefix)

	var lastHeight uint64
	for {
		lastHeight = s.scanNewHeight(lastHeight)
	}
}

func (s *BlockScanner) scanNewHeight(lastHeight uint64) uint64 {
	s.adaptor.Transfers(4345674)

	return 0
	lastScannerHeight, err := s.adaptor.LastHeight()
	if err != nil {
		log.Errorf("%s: Error: %s", s.prefix, err.Error())
		return lastHeight
	} else {
		if lastHeight >= lastScannerHeight {
			return lastHeight
		}

		log.Infof("%s: Scan new block: %d", s.prefix, lastScannerHeight)
		err := s.scanBlock(lastScannerHeight)
		if err != nil {
			log.Errorf("%s: Error scan block: %s", s.prefix, err.Error())
		}
		return lastScannerHeight
	}
}

func (s *BlockScanner) scanBlock(number uint64) error {
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
						log.Errorf("invalid get by address in notification service: %s", err.Error())
						continue
					}
					if err != db.ErrNoRows {
						name = p.Name
					}

					msg := s.notificator.Msg(name, firebase.Sent, currency.ConvertFromPlanck(amount), currency)
					err = s.notificator.Notify(msg, sub.Token)

					log.Infof("%s: Notify Type: Sent; Sender:%s; Receiver:%s; Sub:%s Amount:%s;",
						s.prefix, sub.Address, receiver, sub.Address, amount.String())
					if err != nil {
						log.Error("%s: Error: %s", s.prefix, err.Error())
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
						log.Infof("invalid get by address in notification service: %s", err.Error())
						continue
					}
					if err != db.ErrNoRows {
						name = p.Name
					}

					msg := s.notificator.Msg(name, firebase.Received, currency.ConvertFromPlanck(amount), currency)
					err = s.notificator.Notify(msg, sub.Token)

					log.Infof("%s: Notify Type: Received; Sender:%s; Receiver:%s; Sub:%s Amount:%s;",
						s.prefix, sender, sub.Address, sub.Address, amount.String())
					if err != nil {
						log.Errorf("%s: Error: %s", s.prefix, err.Error())
					}
				}
			}
			delete(receivers, sub.Address)
		}
	}

	return nil
}

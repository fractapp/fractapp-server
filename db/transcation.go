package db

import (
	"fractapp-server/types"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

type TxAction int32
type Status int32
type TxDirection int32

const (
	Transfer TxAction = iota
	StakingReward
	StakingCreateWithdrawalRequest
	StakingWithdrawn
	StakingOpenDeposit
	StakingAddAmount
	ConfirmWithdrawal
	UpdateNomination
)

const (
	Success Status = iota
	Fail
)

const (
	NoneDirection TxDirection = iota
	OutDirection
	InDirection
)

type Transaction struct {
	Id            ID             `bson:"_id"`
	TxId          string         `bson:"tx_id"`
	Hash          string         `bson:"hash"`
	Currency      types.Currency `bson:"currency"`
	MemberAddress string         `bson:"member_address"`
	MemberId      *ID            `bson:"member_id"`
	Owner         ID             `bson:"owner"`
	Direction     TxDirection    `bson:"direction"`
	Action        TxAction       `bson:"action"`
	Status        Status         `bson:"status"`
	Value         string         `bson:"value"`
	Fee           string         `bson:"fee"`
	Price         float32        `bson:"price"`
	Timestamp     int64          `bson:"timestamp"`
}

func (db *MongoDB) TransactionById(id ID) (*Transaction, error) {
	tx := &Transaction{}

	collection := db.collections[TransactionsDB]
	res := collection.FindOne(db.ctx, bson.D{
		{"_id", id},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(tx)
	if err != nil {
		return nil, err
	}

	return tx, err
}

func (db *MongoDB) TransactionByTxIdAndOwner(txId string, owner ID) (*Transaction, error) {
	tx := &Transaction{}

	collection := db.collections[TransactionsDB]
	res := collection.FindOne(db.ctx, bson.D{
		{"tx_id", txId},
		{"owner", owner},
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}

	err = res.Decode(tx)
	if err != nil {
		return nil, err
	}

	return tx, err
}

func (db *MongoDB) TransactionsByOwner(ownerAddress string, currency types.Currency) ([]Transaction, error) {
	collection := db.collections[TransactionsDB]
	transactions := make([]Transaction, 0)
	opt := options.Find()
	opt.SetSort(bson.D{{"timestamp", 1}})

	res, err := collection.Find(db.ctx, bson.D{
		{"currency", currency},
		{"$or", []interface{}{
			bson.D{{"from", ownerAddress}},
			bson.D{{"to", ownerAddress}},
		}},
	}, opt)
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, err
}

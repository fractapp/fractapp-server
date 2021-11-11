package subscriber

import (
	"encoding/json"
	"fractapp-server/controller"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	"fractapp-server/push"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const NotifyRoute = "/notify"

type Controller struct {
	db db.DB
}

func NewController(db db.DB) *Controller {
	return &Controller{
		db: db,
	}
}

func (c *Controller) MainRoute() string {
	return "/"
}

func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case NotifyRoute:
		return c.notifyRoute, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	default:
		log.Errorf("Error: %d", err)
		http.Error(w, "", http.StatusBadRequest)
	}
}

func (c *Controller) notifyRoute(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	log.Println("txs: " + string(b))

	txs := make([]profile.Transaction, 0)
	err = json.Unmarshal(b, &txs)
	if err != nil {
		return err
	}

	for _, v := range txs {
		currency := v.Currency
		txTime := time.Unix(v.Timestamp/1000, 0)
		prices, err := c.db.Prices(currency.String(), txTime.
			Add(-15*time.Minute).Unix()*1000, txTime.
			Add(15*time.Minute).Unix()*1000)

		if err != nil {
			return err
		}

		price := float32(0)
		// search for a value with a minimum difference from the transaction time
		if len(prices) > 0 {
			price = prices[0].Price
			diff := v.Timestamp - prices[0].Timestamp
			for _, p := range prices {
				newDiff := v.Timestamp - p.Timestamp
				if newDiff < 0 {
					newDiff *= -1
				}

				if newDiff < diff {
					diff = newDiff
					price = p.Price
				}
			}
		}

		var senderId *db.ID
		senderProfile, err := c.db.ProfileByAddress(currency.Network(), v.From)
		if err != nil && err != db.ErrNoRows {
			return err
		}
		if senderProfile != nil {
			senderId = &senderProfile.Id
		}

		var receiverId *db.ID
		receiverProfile, err := c.db.ProfileByAddress(currency.Network(), v.To)
		if err != nil && err != db.ErrNoRows {
			return err
		}
		if receiverProfile != nil {
			receiverId = &receiverProfile.Id
		}

		var senderTx *db.Transaction
		dbTxs := make([]*db.Transaction, 0)
		if v.Action == db.Transfer {
			if senderId != nil {
				senderTx = &db.Transaction{
					Id:            db.NewId(),
					TxId:          v.ID,
					Hash:          v.Hash,
					Currency:      v.Currency,
					MemberAddress: v.To,
					MemberId:      receiverId,
					Owner:         *senderId,
					Direction:     db.OutDirection,
					Status:        v.Status,
					Value:         v.Value,
					Fee:           v.Fee,
					Price:         price,
					Timestamp:     v.Timestamp,
				}

				dbTxs = append(dbTxs, senderTx)
			}
		}

		var receiverTx *db.Transaction
		if receiverId != nil {
			receiverTx = &db.Transaction{
				Id:            db.NewId(),
				TxId:          v.ID,
				Hash:          v.Hash,
				Currency:      v.Currency,
				MemberAddress: v.From,
				MemberId:      senderId,
				Owner:         *receiverId,
				Direction:     db.InDirection,
				Status:        v.Status,
				Value:         v.Value,
				Fee:           v.Fee,
				Price:         price,
				Timestamp:     v.Timestamp,
			}

			dbTxs = append(dbTxs, receiverTx)
		}

		for _, dbTx := range dbTxs {
			_, err = c.db.TransactionByTxIdAndOwner(dbTx.TxId, dbTx.Owner)
			if err != nil && err != db.ErrNoRows {
				return err
			} else if err == nil { //if tx is exist in db
				continue
			}

			err = c.db.Insert(dbTx)
			if err != nil {
				return err
			}
		}

		if v.Action != db.Transfer && v.Action != db.StakingReward {
			continue
		}

		amount, _ := new(big.Int).SetString(v.Value, 10)
		fAmount, _ := currency.ConvertFromPlanck(amount).Float64()
		usdAmount := fAmount * float64(price)

		if v.Action == db.Transfer {
			senderTitle := v.From
			if senderProfile != nil {
				if senderProfile.Name != "" {
					senderTitle = senderProfile.Name
				} else {
					senderTitle = "@" + senderProfile.Username
				}
			}

			receiverTitle := v.To
			if receiverProfile != nil {
				if receiverProfile.Name != "" {
					receiverTitle = receiverProfile.Name
				} else {
					receiverTitle = "@" + receiverProfile.Username
				}
			}

			notifications := make([]interface{}, 0)
			if senderTx != nil && senderProfile != nil {
				notifications = append(notifications, &db.Notification{
					Id:        db.NewId(),
					Title:     receiverTitle,
					Message:   push.CreateMsg(push.Sent, fAmount, usdAmount, currency),
					Type:      db.TransactionNotificationType,
					TargetId:  senderTx.Id,
					UserId:    senderProfile.Id,
					Timestamp: time.Now().Unix(),
				})
			}

			if receiverTx != nil && receiverProfile != nil {
				notifications = append(notifications, db.Notification{
					Id:        db.NewId(),
					Title:     senderTitle,
					Message:   push.CreateMsg(push.Received, fAmount, usdAmount, currency),
					Type:      db.TransactionNotificationType,
					TargetId:  receiverTx.Id,
					UserId:    receiverProfile.Id,
					Timestamp: time.Now().Unix(),
				})
			}

			if len(notifications) > 0 {
				err = c.db.InsertMany(notifications)
				if err != nil {
					return err
				}
			}
		} else if v.Action == db.StakingReward && receiverTx != nil && receiverProfile != nil {
			notification := &db.Notification{
				Id:        db.NewId(),
				Title:     "Deposit payout",
				Message:   push.CreateMsg(push.Received, fAmount, usdAmount, currency),
				Type:      db.TransactionNotificationType,
				TargetId:  receiverTx.Id,
				UserId:    receiverProfile.Id,
				Timestamp: time.Now().Unix(),
			}
			err = c.db.Insert(notification)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

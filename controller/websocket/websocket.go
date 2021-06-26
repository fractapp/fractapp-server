package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	"fractapp-server/types"
	"fractapp-server/utils"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/jwtauth"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	connectionClosedErr = errors.New("ws connection closed")
)

func CreateConnectRoute(jwtAuth *jwtauth.JWTAuth, authMiddleware *middleware.AuthMiddleware, pgDB db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := connect(jwtAuth, authMiddleware, w, r, pgDB)
		if err == middleware.InvalidAuthErr {
			log.Errorf("Error: %d \n", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Errorf("Error: %d \n", err)
			http.Error(w, controller.InvalidAuthErr.Error(), http.StatusBadRequest)
			return
		}
	}
}

var connections sync.Map
var connectionsMutex sync.Map

func connect(jwtAuth *jwtauth.JWTAuth, authMiddleware *middleware.AuthMiddleware, w http.ResponseWriter, r *http.Request, pgDB db.DB) error {
	id, err := auth(jwtAuth, authMiddleware, r)
	if err != nil {
		return err
	}

	log.Infof("Try connection: %s\n", id)

	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true //TODO
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	defer func() {
		connections.Delete(id)
		c.Close()
	}()

	v, ok := connections.Load(id)
	if ok {
		mu, muOK := connectionsMutex.Load(id)
		if !muOK {
			return nil
		}
		mu.(*sync.Mutex).Lock()
		v.(*websocket.Conn).Close()
		mu.(*sync.Mutex).Unlock()
	} else {
		connectionsMutex.Store(id, &sync.Mutex{})
	}
	connections.Store(id, c)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		log.Infof("id: %s; message: %s\n", id, string(message))

		rq := &Request{}
		err = json.Unmarshal(message, rq)
		if err != nil {
			log.Infof("id: %s; err: %s\n", id, err)
			continue
		}

		response := Response{
			Id:     rq.Id,
			Method: rq.Method,
			Ok:     false,
		}

		switch rq.Method {
		case Messages:
			err = writeMsg(id, rq.Message, pgDB)
			/*var dbMsgs []db.Message //TODO: my messages for history and is_delivered
			dbMsgs, err = pgDB.MessagesByReceiver(id)
			if err != nil {
				break
			}

			users := make(map[string]profile.ShortUserProfile)
			messages := make([]MessageRs, 0)
			for _, msg := range dbMsgs {
				messages = append(messages, MessageRs{
					Id:        msg.Id,
					Version:   0,
					Value:     msg.Value,
					Rows:      nil,
					Sender:    msg.SenderId,
					Timestamp: msg.Timestamp,
				})

				if _, isOK := users[msg.SenderId]; !isOK {
					user, err := userById(msg.SenderId, pgDB)
					if err != nil {
						break
					}
					users[msg.SenderId] = *user
				}
			}
			response.Value = &MessagesAndUsersRs{
				Messages: messages,
				Users:    users,
			}*/
		}

		if err == nil {
			response.Ok = true
		} else {
			response.Ok = false
		}

		rBytes, _ := json.Marshal(response)
		log.Infof("id: %s; resonse: %s\n", id, rBytes)

		if !response.Ok {
			log.Infof("id: %s; err (response): %s\n", id, err)
		}

		err = sendWsData(response, id)
		if err != nil {
			return err
		}
	}
}

func userById(id string, pgDB db.DB) (*profile.ShortUserProfile, error) {
	user, err := pgDB.ProfileById(id)
	if err != nil {
		return nil, err
	}

	addresses, err := pgDB.AddressesById(user.Id)
	if err != nil {
		return nil, err
	}

	p := profile.ShortUserProfile{
		Id:         user.Id,
		Name:       user.Name,
		Username:   user.Username,
		AvatarExt:  user.AvatarExt,
		LastUpdate: user.LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}

	for _, v := range addresses {
		p.Addresses[v.Network.Currency()] = v.Address
	}

	return &p, nil
}
func writeMsg(sender string, msg *MessageRq, pgDB db.DB) error {
	if msg == nil {
		return errors.New("invalid value") //TODO: migrate to var
	}

	if sender == msg.Receiver {
		return errors.New("invalid receiver")
	}

	senderProfile, err := pgDB.ProfileById(sender)
	if err != nil {
		return err
	}

	receiverProfile, err := pgDB.ProfileById(msg.Receiver)
	if err != nil {
		return err
	}

	if (!senderProfile.IsChatBot && !receiverProfile.IsChatBot) ||
		(senderProfile.IsChatBot && receiverProfile.IsChatBot) || (sender == msg.Receiver) {
		return errors.New("invalid receiver")
	}

	id, err := utils.RandomHex(32)
	if err != nil {
		return err
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	dbMessage := &db.Message{
		Id:          id,
		Value:       msg.Value,
		SenderId:    sender,
		ReceiverId:  msg.Receiver,
		Timestamp:   timestamp,
		IsDelivered: false,
	}
	err = pgDB.Insert(dbMessage)
	if err != nil {
		return err
	}

	if strings.HasPrefix(msg.Value, "/") {

		switch Action(msg.Value[1:]) {
		case AddTxToChat:

		}
		address := r.URL.Query().Get("address")
		currencyInt, err := strconv.ParseInt(r.URL.Query().Get("currency"), 10, 32)
		if err != nil {
			return err
		}
		currency := types.Currency(currencyInt)

		resp, err := http.Get(fmt.Sprintf("%s/transactions/%s?currency=%s", c.txApiHost, address, currency.String()))
		if err != nil {
			return InvalidConnectionTxApiErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return InvalidConnectionTxApiErr
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		txs := make([]types.Transaction, 0)
		err = json.Unmarshal(body, &txs)
		if err != nil {
			return err
		}

		responseTxs := make([]Transaction, 0)
		for _, v := range txs {
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

			value, _ := new(big.Int).SetString(v.Value, 10)
			fee, _ := new(big.Int).SetString(v.Fee, 10)

			floatValue := currency.ConvertFromPlanck(value)
			floatFee := currency.ConvertFromPlanck(fee)

			a := currency.Accuracy()
			aBig := new(big.Float).SetInt(big.NewInt(a))

			fv, _ := new(big.Float).Mul(floatValue, aBig).Float32()
			fv /= float32(a)
			usdValue := price * fv

			ff, _ := new(big.Float).Mul(floatFee, aBig).Float32()
			ff /= float32(a)
			usdFee := price * ff

			userFrom := ""
			p, err := c.db.ProfileByAddress(v.From)
			if err != nil && err != db.ErrNoRows {
				return err
			}
			if p != nil {
				userFrom = p.Id
			}

			userTo := ""
			p, err = c.db.ProfileByAddress(v.To)
			if err != nil && err != db.ErrNoRows {
				return err
			}
			if p != nil {
				userTo = p.Id
			}

			responseTxs = append(responseTxs, Transaction{
				ID:   v.ID,
				Hash: v.Hash,

				Currency: v.Currency,

				From:     v.From,
				UserFrom: userFrom,

				To:     v.To,
				UserTo: userTo,

				Value:      v.Value,
				UsdValue:   usdValue,
				FloatValue: floatValue.String(),

				Fee:       v.Fee,
				UsdFee:    usdFee,
				FloatFee:  floatFee.String(),
				Timestamp: v.Timestamp,
				Status:    v.Status,
			})
		}

		rsByte, err := json.Marshal(responseTxs)
		if err != nil {
			return err
		}
	}

	msgs, err := pgDB.MessagesBySenderAndReceiver(dbMessage.SenderId, dbMessage.ReceiverId)
	if err != nil {
		return err
	}

	users := make(map[string]profile.ShortUserProfile)
	messages := make([]MessageRs, 0)
	for _, msg := range msgs {
		messages = append(messages, MessageRs{
			Id:        msg.Id,
			Version:   0,
			Value:     msg.Value,
			Rows:      nil,
			Sender:    msg.SenderId,
			Timestamp: msg.Timestamp,
		})

		if _, isOK := users[msg.SenderId]; !isOK {
			user, err := userById(msg.SenderId, pgDB)
			if err != nil {
				break
			}
			users[msg.SenderId] = *user
		}
	}

	err = sendWsData(Response{
		Method: Messages,
		Value: &MessagesAndUsersRs{
			Messages: messages,
			Users:    users,
		},
		Ok: true,
	}, msg.Receiver)

	if err == nil {
		for _, msg := range msgs {
			err := pgDB.UpdateDeliveredMessage(msg.Id, msg.ReceiverId)
			if err != nil {
				return err
			}
		}
	} else {
		//TODO: notification to firebase
	}

	return nil
}

func sendWsData(data interface{}, id string) error {
	c, cOk := connections.Load(id)
	if !cOk {
		return connectionClosedErr
	}

	mu, muOK := connectionsMutex.Load(id)
	if !muOK {
		return nil
	}

	mu.(*sync.Mutex).Lock()
	err := c.(*websocket.Conn).WriteJSON(data)
	mu.(*sync.Mutex).Unlock()

	return err
}

func auth(jwtAuth *jwtauth.JWTAuth, authMiddleware *middleware.AuthMiddleware, r *http.Request) (string, error) {
	token, err := jwtauth.VerifyRequest(jwtAuth, r, jwtauth.TokenFromQuery)
	if err != nil {
		return "", err
	}

	ctx := jwtauth.NewContext(r.Context(), token, nil)
	id, err := authMiddleware.AuthWithJwt(r.WithContext(ctx), jwtauth.TokenFromQuery)
	if err != nil {
		return "", err
	}

	return id, nil
}

package websocket

import (
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/controller/info"
	"fractapp-server/controller/message"
	"fractapp-server/controller/middleware"
	"fractapp-server/controller/profile"
	"fractapp-server/controller/substrate"
	"fractapp-server/db"
	"fractapp-server/types"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/go-chi/jwtauth"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const ConnectRoute = "/connect"

var (
	connectionClosedErr = errors.New("ws connection closed")
)

type (
	UserData struct {
		Conn  *websocket.Conn
		Mutex *sync.Mutex
	}

	Controller struct {
		db             db.DB
		jwtAuth        *jwtauth.JWTAuth
		authMiddleware *middleware.AuthMiddleware
		txApiHost      string
		connections    sync.Map
	}
)

func NewController(
	db db.DB,
	jwtAuth *jwtauth.JWTAuth,
	authMiddleware *middleware.AuthMiddleware,
	txApiHost string,
) *Controller {
	return &Controller{
		db:             db,
		jwtAuth:        jwtAuth,
		authMiddleware: authMiddleware,
		txApiHost:      txApiHost,
	}
}

func (c *Controller) MainRoute() string {
	return "/ws"
}
func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case ConnectRoute:
		return c.connect, nil
	}

	return nil, controller.InvalidRouteErr
}

func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case middleware.InvalidAuthErr:
		log.Errorf("Ws error: %d \n", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	default:
		log.Errorf("Ws error: %d \n", err)
		http.Error(w, controller.InvalidAuthErr.Error(), http.StatusBadRequest)
	}
}

func (c *Controller) connect(w http.ResponseWriter, r *http.Request) error {
	authId, profileId, err := c.auth(r)
	if err != nil {
		return err
	}

	log.Infof("Try connection: %s\n", authId)

	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true //TODO: security?
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	defer func() {
		connection.Close()
		c.connections.Delete(authId)
	}()

	v, ok := c.connections.Load(authId)
	if ok {
		userData := v.(*UserData)
		userData.Mutex.Lock()
		userData.Conn.Close()
		userData.Mutex.Unlock()
		c.connections.Delete(authId)
	}

	c.connections.Store(authId, &UserData{
		Conn:  connection,
		Mutex: &sync.Mutex{},
	})

	userProfile, err := c.db.ProfileById(profileId)
	if err != nil {
		return err
	}

	q := make(chan bool)
	go c.scheduler(q, userProfile)

	defer func() {
		q <- true
	}()

	for {
		log.Infof("ws - id: %s; \n", authId)

		_, b, err := connection.ReadMessage()
		if err != nil {
			log.Errorf("ws - id: %s; error: %s\n", authId, err.Error())
			return nil
		}

		rq := &Rq{}
		err = json.Unmarshal(b, rq)
		if err != nil {
			log.Errorf("ws - id: %s; error: %s\n", authId, err.Error())
			continue
		}

		var v interface{}
		switch rq.Method {
		case setDeliveredMethod:
			c.setDelivered(rq, userProfile)
		case getTxsStatusesMethod:
			v = c.getTxsStatuses(rq, authId)
		case getUsersMethod:
			v = c.getUsers(rq)
		}

		if v != nil {
			err = c.SendWsData(v, authId)
			if err != nil {
				log.Errorf("ws - id: %s; error: %s\n", authId, err.Error())
				return err
			}
		}
	}

	return nil
}

func (c *Controller) getUsers(rq *Rq) *WsResponse {
	usersProfiles := make(map[string]*profile.ShortUserProfile)
	for _, authId := range rq.Ids {
		p, err := c.db.ProfileByAuthId(authId)
		if err != nil {
			log.Errorf("ws - id: %s; error: %s\n", authId, err.Error())
			continue
		}

		shortProfile := &profile.ShortUserProfile{
			Id:         p.AuthId,
			Name:       p.Name,
			Username:   p.Username,
			AvatarExt:  p.AvatarExt,
			LastUpdate: p.LastUpdate,
			IsChatBot:  p.IsChatBot,
			Addresses:  make(map[types.Network]string),
		}

		for k, v := range p.Addresses {
			shortProfile.Addresses[k] = v.Address
		}

		usersProfiles[p.AuthId] = shortProfile
	}

	return &WsResponse{
		Method: usersMethod,
		Value:  usersProfiles,
	}
}

func (c *Controller) getTxsStatuses(rq *Rq, authId string) *WsResponse {
	txsStatuses := make([]*profile.TxStatusRs, 0)
	for _, txHash := range rq.Ids {
		status, err := profile.TxStatus(c.txApiHost, txHash)
		if err != nil {
			log.Errorf("ws - id: %s; error: %s\n", authId, err.Error())
			continue
		}
		txsStatuses = append(txsStatuses, status)
	}

	return &WsResponse{
		Method: txsStatusesMethod,
		Value:  txsStatuses,
	}
}

func (c *Controller) setDelivered(rq *Rq, userProfile *db.Profile) {
	deliveredMap := make(map[string]bool)
	for _, id := range rq.Ids {
		deliveredMap[id] = true
	}

	notifications, err := c.db.UndeliveredNotificationsByUserId(userProfile.Id)
	if err != nil {
		log.Errorf("ws - id: %s; error: %s\n", userProfile.AuthId, err.Error())
		return
	}

	for _, notification := range notifications {
		stringId := primitive.ObjectID(notification.Id).Hex()
		if _, ok := deliveredMap[stringId]; ok {
			notification.Delivered = true
			err := c.db.UpdateByPK(notification.Id, &notification)
			if err != nil {
				log.Errorf("ws - id: %s; error: %s\n", userProfile.AuthId, err.Error())
				continue
			}
		}
	}
}

func (c *Controller) scheduler(q chan bool, user *db.Profile) {
	for {
		select {
		case <-q:
			log.Errorf("ws - exit ws sheduler: %s; \n", user.AuthId)
			return
		default:
			go c.notifications(user)
			go c.balances(user)
			time.Sleep(time.Second)
		}
	}
}

func (c *Controller) notifications(user *db.Profile) {
	transactionsByCurrency := make(map[types.Currency][]*message.TransactionRs)

	notifications, err := c.db.UndeliveredNotificationsByUserId(user.Id)
	usersById := make(map[db.ID]db.Profile)
	messagesRs := make([]*message.MessageRs, 0)

	deliveredNotifications := make([]string, 0)

	for _, notification := range notifications {
		var dbMsg *db.Message
		var dbTx *db.Transaction
		var memberId *db.ID

		if notification.Type == db.MessageNotificationType {
			dbMsg, err = c.db.MessageById(notification.TargetId)
			if err != nil && err != db.ErrNoRows {
				log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
				continue
			} else if err == db.ErrNoRows {
				notification.Delivered = true
				err := c.db.UpdateByPK(notification.Id, &notification)
				if err != nil {
					log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
				}
				continue
			}
			memberId = &dbMsg.SenderId
		} else if notification.Type == db.TransactionNotificationType {
			dbTx, err = c.db.TransactionById(notification.TargetId)
			if err != nil && err != db.ErrNoRows {
				log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
				continue
			} else if err == db.ErrNoRows {
				notification.Delivered = true
				err := c.db.UpdateByPK(notification.Id, &notification)
				if err != nil {
					log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
				}
				continue
			}

			memberId = dbTx.MemberId
		}

		if memberId != nil {
			if _, ok := usersById[*memberId]; !ok {
				notificationUser, err := c.db.ProfileById(*memberId)
				if err != nil {
					log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
					continue
				}

				usersById[*memberId] = *notificationUser
			}
		}

		if dbTx != nil {
			var authId *string
			if memberId != nil {
				id := usersById[*memberId].AuthId
				authId = &id
			}

			transactionsByCurrency[dbTx.Currency] = append(transactionsByCurrency[dbTx.Currency], &message.TransactionRs{
				Id:            dbTx.TxId,
				Hash:          dbTx.Hash,
				Currency:      dbTx.Currency,
				MemberAddress: dbTx.MemberAddress,
				Member:        authId,
				Direction:     dbTx.Direction,
				Action:        dbTx.Action,
				Status:        dbTx.Status,
				Value:         dbTx.Value,
				Fee:           dbTx.Value,
				Price:         dbTx.Price,
				Timestamp:     dbTx.Timestamp,
			})
		}

		if dbMsg != nil && memberId != nil {
			sender := usersById[*memberId]
			messagesRs = append(messagesRs, &message.MessageRs{
				Id:        primitive.ObjectID(dbMsg.Id).Hex(),
				Args:      dbMsg.Args,
				Action:    message.Action(dbMsg.Action),
				Version:   dbMsg.Version,
				Value:     dbMsg.Value,
				Rows:      dbMsg.Rows,
				Sender:    sender.AuthId,
				Receiver:  user.AuthId,
				Timestamp: dbMsg.Timestamp,
			})
		}

		deliveredNotifications = append(deliveredNotifications, primitive.ObjectID(notification.Id).Hex())
	}

	users := make(map[string]profile.ShortUserProfile)
	for _, user := range usersById {
		p := profile.ShortUserProfile{
			Id:         user.AuthId,
			Name:       user.Name,
			Username:   user.Username,
			AvatarExt:  user.AvatarExt,
			LastUpdate: user.LastUpdate,
			IsChatBot:  user.IsChatBot,
			Addresses:  make(map[types.Network]string),
		}

		for k, v := range user.Addresses {
			p.Addresses[k] = v.Address
		}

		users[user.AuthId] = p
	}

	prices := make([]*info.Price, 0)
	for _, v := range types.Currencies {
		price, err := c.db.LastPriceByCurrency(v.String())
		if err != nil && err != db.ErrNoRows {
			log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
			continue
		}
		if err == db.ErrNoRows {
			continue
		}

		prices = append(prices, &info.Price{
			Currency: v,
			Value:    price.Price,
		})
	}

	err = c.SendWsData(&WsResponse{
		Method: updateMethod,
		Value: &Update{
			Messages:      messagesRs,
			Transactions:  transactionsByCurrency,
			Users:         users,
			Notifications: deliveredNotifications,
			Prices:        prices,
		},
	}, user.AuthId)
	if err != nil {
		log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
	}
}

func (c *Controller) balances(user *db.Profile) {
	balanceByCurrency := make(map[types.Currency]*substrate.Balance)

	for network, value := range user.Addresses {
		currency := network.Currency()
		balance, err := substrate.SubstrateBalance(c.txApiHost, value.Address, currency)
		if err != nil {
			log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
			continue
		}

		balanceByCurrency[currency] = balance
	}

	err := c.SendWsData(&WsResponse{
		Method: balancesMethod,
		Value: &Balances{
			Balances: balanceByCurrency,
		},
	}, user.AuthId)
	if err != nil {
		log.Errorf("ws - id: %s; error: %s\n", user.AuthId, err.Error())
	}
}

func (c *Controller) SendWsData(data interface{}, id string) error {
	userData, cOk := c.connections.Load(id)
	if !cOk {
		return connectionClosedErr
	}

	user := userData.(*UserData)
	user.Mutex.Lock()
	err := user.Conn.WriteJSON(data)
	user.Mutex.Unlock()

	b, _ := json.Marshal(data)
	log.Infof("ws - send data (%s): %s; \n", id, b)

	return err
}

func (c *Controller) auth(r *http.Request) (string, db.ID, error) {
	token, err := jwtauth.VerifyRequest(c.jwtAuth, r, jwtauth.TokenFromQuery)
	if err != nil {
		return "", db.ID{}, err
	}

	ctx := jwtauth.NewContext(r.Context(), token, nil)
	authId, profileId, err := c.authMiddleware.AuthWithJwt(r.WithContext(ctx), jwtauth.TokenFromQuery)
	if err != nil {
		return "", db.ID{}, err
	}

	return authId, profileId, nil
}

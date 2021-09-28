package message

import (
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	"fractapp-server/push"
	"fractapp-server/types"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UnreadRoute = "/unread"
	SendRoute   = "/send"
	ReadRoute   = "/read"
)

type Controller struct {
	notificator push.Notificator
	db          db.DB
}

func NewController(db db.DB, notificator push.Notificator) *Controller {
	return &Controller{
		notificator: notificator,
		db:          db,
	}
}

func (c *Controller) MainRoute() string {
	return "/message"
}
func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case UnreadRoute:
		return c.unread, nil
	case SendRoute:
		return c.send, nil
	case ReadRoute:
		return c.read, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case db.ErrNoRows:
		http.Error(w, "", http.StatusNotFound)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// unread godoc
// @Summary Unread messages
// @Description get unread messages
// @ID unread
// @Tags Message
// @Accept  json
// @Produce json
// @Success 200 {object} MessagesAndUsersRs
// @Failure 400 {string} string
// @Router /message/unread [get]
func (c *Controller) unread(w http.ResponseWriter, r *http.Request) error {
	id := middleware.AuthId(r)
	receiverProfile, err := c.db.ProfileByAuthId(id)
	if err != nil {
		return err
	}

	dbMessages, err := c.db.MessagesByReceiver(receiverProfile.Id)
	if err != nil {
		return err
	}

	usersById := make(map[db.ID]db.Profile)
	messages := make([]MessageRs, 0)
	for _, msg := range dbMessages {
		if _, ok := usersById[msg.SenderId]; !ok {
			user, err := c.db.ProfileById(msg.SenderId)
			if err != nil {
				return err
			}

			usersById[user.Id] = *user
		}

		sender := usersById[msg.SenderId]

		messages = append(messages, MessageRs{
			Id:        primitive.ObjectID(msg.Id).Hex(),
			Args:      msg.Args,
			Action:    Action(msg.Action),
			Version:   msg.Version,
			Value:     msg.Value,
			Rows:      msg.Rows,
			Sender:    sender.AuthId,
			Receiver:  receiverProfile.AuthId,
			Timestamp: msg.Timestamp,
		})
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

	err = controller.JSON(w, &MessagesAndUsersRs{
		Messages: messages,
		Users:    users,
	})
	if err != nil {
		return err
	}

	return nil
}

// read godoc
// @Summary Read messages
// @Description read messages
// @ID read
// @Tags Message
// @Accept  json
// @Produce json
// @Param rq body []string true "array of message ids"
// @Success 200
// @Failure 400 {string} string
// @Router /message/read [post]
func (c *Controller) read(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := middleware.ProfileId(r)

	rq := make([]string, 0)
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	for _, stringMsgId := range rq {
		msgId, err := primitive.ObjectIDFromHex(stringMsgId)
		if err != nil {
			continue
		}

		err = c.db.SetDelivered(db.ID(id), db.ID(msgId))
		if err != nil {
			continue
		}
	}

	return nil
}

// sendMsg godoc
// @Summary send message
// @Description send message
// @ID read
// @Tags Message
// @Accept  json
// @Produce json
// @Param rq body MessageRq true "send message body"
// @Success 200 {object} SendInfo
// @Failure 400 {string} string
// @Router /message/read [post]
func (c *Controller) send(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	senderId := middleware.AuthId(r)

	msg := MessageRq{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		return err
	}

	if senderId == msg.Receiver {
		return errors.New("invalid receiver")
	}

	senderProfile, err := c.db.ProfileByAuthId(senderId)
	if err != nil {
		return err
	}

	receiverProfile, err := c.db.ProfileByAuthId(msg.Receiver)
	if err != nil {
		return err
	}

	if (!senderProfile.IsChatBot && !receiverProfile.IsChatBot) ||
		(senderProfile.IsChatBot && receiverProfile.IsChatBot) || (senderId == msg.Receiver) {
		return errors.New("invalid receiver")
	}

	if !senderProfile.IsChatBot && (msg.Rows != nil || len(msg.Rows) != 0) {
		return errors.New("invalid msg")
	}

	senderTitle := "@" + senderProfile.Username
	if senderProfile.Name != "" {
		senderTitle = senderProfile.Name
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	dbMessage := &db.Message{
		Id:          db.NewId(),
		Value:       msg.Value,
		Action:      msg.Action,
		Version:     msg.Version,
		Args:        msg.Args,
		Rows:        msg.Rows,
		SenderId:    senderProfile.Id,
		ReceiverId:  receiverProfile.Id,
		Timestamp:   timestamp,
		IsDelivered: false,
	}

	err = c.db.Insert(dbMessage)
	if err != nil {
		return err
	}

	if !receiverProfile.IsChatBot {
		token, err := c.db.TokenByProfileId(receiverProfile.Id)
		if err != nil {
			log.Errorf("Push notification error: %d \n", err)
		}

		notifyErr := c.notificator.Notify(msg.Value, senderTitle, token.Token)
		if notifyErr != nil {
			log.Errorf("Push notification error: %d \n", notifyErr)
		}
	}

	err = controller.JSON(w, &SendInfo{
		Timestamp: timestamp,
	})
	if err != nil {
		return err
	}

	return nil
}

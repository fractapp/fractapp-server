package messages

import (
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"net/http"
	"strconv"
	"strings"
)

const (
	UpdatesRoute = "/updates"
	SendRoute    = "/send"
)

var (
	InvalidFileFormatErr = errors.New("invalid file format")
)

type Controller struct {
	db db.DB
}

func NewController(db db.DB) *Controller {
	return &Controller{
		db: db,
	}
}

func (c *Controller) MainRoute() string {
	return "/messages"
}
func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case UpdatesRoute:
		return c.search, nil
	case SendRoute:
		return c.updateProfile, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case db.ErrNoRows:
		http.Error(w, "", http.StatusNotFound)
	case InvalidFileFormatErr:
		fallthrough
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// updates godoc
// @Summary Get new messages
// @Description get new messages by account
// @ID search
// @Tags Messages
// @Accept  json
// @Produce json
// @Param after query int true "get messages after timestamp"
// @Success 200 {object} []Message
// @Failure 400 {string} string
// @Failure 404
// @Router /messages/updates [get]
func (c *Controller) updates(w http.ResponseWriter, r *http.Request) error {
	id := middleware.AuthId(r)
	timestampStr := strings.Trim(r.URL.Query().Get("timestamp"), " ")

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return err
	}

	messages, err := c.db.GetMessagesByReceiver(id, timestamp)
	if err != nil {
		return err
	}

	response := make([]Message, 0)

	for _, msg := range messages {
		response = append(response, Message{
			Id:        msg.Id,
			Value:     msg.Value,
			Buttons:   msg.Buttons,
			SenderId:  msg.SenderId,
			Timestamp: msg.Timestamp,
		})
	}

	b, err := json.Marshal(&response)
	if err != nil {
		return err
	}

	w.Write(b)
	return nil
}

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fractapp-server/config"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"fractapp-server/types"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/go-pg/pg/v10"
	log "github.com/sirupsen/logrus"
)

var (
	host       = "127.0.0.1:9505"
	configPath = "config.json"
)

type NotifierRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Currency int    `json:"currency"`
}

func init() {
	flag.StringVar(&host, "host", host, "host for server")
	flag.StringVar(&configPath, "config", configPath, "config file")
	flag.Parse()
}

var database *db.PgDB
var notificator firebase.TxNotificator

func main() {
	log.Info("Start price cache ...")

	ctx, cancel := context.WithCancel(context.Background())
	config, err := config.Parse(configPath)
	if err != nil {
		log.Fatalf("Invalid parse config: %s", err.Error())
	}

	notificator, err = firebase.NewClient(ctx, "firebase.json", config.Firebase.ProjectId)
	if err != nil {
		log.Fatalf("Invalid create notificator: %s", err.Error())
	}

	// connect to db
	pgDb := pg.Connect(&pg.Options{
		Addr:     config.DB.Host,
		User:     config.DB.User,
		Password: config.DB.Password,
		Database: config.DB.Database,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err := pgDb.Ping(ctx); err != nil {
		log.Fatalf("Invalid connect to db: %s", err.Error())
	}

	database = (*db.PgDB)(pgDb)
	// create http server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Group(func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Post("/notify", handler)
		})
	})

	srv := &http.Server{
		Addr:    host,
		Handler: r,
	}

	// start http server
	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("http: Server listen: %s", host)

	// await exit signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	exitCtx, shutDownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutDownCancel()
	srv.Shutdown(exitCtx)

	cancel()
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := notifyRoute(w, r)
	if err != nil {
		log.Errorf("Error: %d \n", err)
		http.Error(w, "", http.StatusBadRequest)
	}
}

func notifyRoute(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var nRq NotifierRequest
	err = json.Unmarshal(b, &nRq)
	if err != nil {
		return err
	}

	currency := types.Currency(nRq.Currency)
	err = sendNotification(nRq.From, nRq.To, firebase.Sent, currency, nRq.Value)
	if err != nil {
		return err
	}

	err = sendNotification(nRq.To, nRq.To, firebase.Received, currency, nRq.Value)
	if err != nil {
		return err
	}

	return nil
}

func sendNotification(ownerAddress string, memberAddress string, txType firebase.TxType, currency types.Currency, value string) error {
	sub, err := database.SubscriberByAddress(ownerAddress)
	if err != nil && err != db.ErrNoRows {
		return err
	}
	if err == db.ErrNoRows {
		return nil
	}

	profile, err := database.ProfileByAddress(memberAddress)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	amount, _ := new(big.Int).SetString(value, 10)
	fAmount, _ := currency.ConvertFromPlanck(amount).Float64()
	if err == db.ErrNoRows || profile == nil {
		msg := notificator.MsgForAddress(memberAddress, txType, fAmount, currency)
		log.Infof("Notify (%s): %s \n", sub.Address, msg)

		notifyErr := notificator.Notify("", msg, sub.Token)
		if notifyErr != nil {
			log.Errorf("%d \n", notifyErr)
		}
	} else {
		msg := notificator.MsgForAuthed(txType, fAmount, currency)
		log.Infof("Notify (%s): %s \n", sub.Address, msg)

		name := ""
		if profile.Name != "" {
			name = profile.Name
		} else {
			name = "@" + profile.Username
		}

		notifyErr := notificator.Notify(name, msg, sub.Token)
		if notifyErr != nil {
			log.Errorf("%d \n", notifyErr)
		}
	}

	return nil
}

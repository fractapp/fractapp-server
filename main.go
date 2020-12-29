package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"fractapp-server/config"
	internalMiddleware "fractapp-server/controller/middleware"
	"fractapp-server/controller/notification"
	"fractapp-server/controller/profile"
	"fractapp-server/notificator"
	"fractapp-server/scanner"
	"fractapp-server/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-pg/pg/v10"
)

var host = "127.0.0.1:9544"
var configPath = "config.json"

func init() {
	flag.StringVar(&host, "host", host, "host for server")
	flag.StringVar(&configPath, "config", configPath, "config file")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	err := start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func start(ctx context.Context) error {
	log.Println("Setup notification service")
	// parse config
	config, err := config.Parse(configPath)
	if err != nil {
		return errors.New(fmt.Sprint("Invalid parse config: ", err.Error()))
	}

	// connect to db
	database := pg.Connect(&pg.Options{
		Addr:     config.DB.Host,
		User:     config.DB.User,
		Password: config.DB.Password,
		Database: config.DB.Database,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	if err := database.Ping(ctx); err != nil {
		return err
	}

	// create http server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(internalMiddleware.PubKeyAuth)

	r.Use(middleware.Timeout(60 * time.Second))

	nController := notification.NewController(database)
	pController := profile.NewController(
		database,
		config.SMSService.FromNumber,
		config.SMSService.AccountSid,
		config.SMSService.AuthToken,
	)

	r.Post("/notification/subscribe", nController.Subscribe)
	r.With(internalMiddleware.PubKeyAuth).Route("/profile", func(r chi.Router) {
		r.Post(string(profile.Auth), pController.Route(profile.Auth))
		r.Post(string(profile.ConfirmAuth), pController.Route(profile.ConfirmAuth))
		r.Post(string(profile.UpdateProfile), pController.Route(profile.UpdateProfile))
		r.Get(string(profile.Username), pController.Route(profile.Username))
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

	n, err := notificator.NewFirebaseNotificator(ctx, config.Firebase.WithCredentialsFile, config.Firebase.ProjectId)
	if err != nil {
		return err
	}
	for k, url := range config.SubstrateUrls {
		network := types.ParseNetwork(k)
		es := scanner.NewEventScanner(url, database, network.String(), network, n)
		go func() {
			err = es.Start()
			if err != nil {
				log.Printf("%s scanner down: %s \n", network.String(), err)
			}
		}()
		log.Printf("Event scanner for %s started \n", k)
	}

	// await exit signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	exitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	srv.Shutdown(exitCtx)

	return nil
}

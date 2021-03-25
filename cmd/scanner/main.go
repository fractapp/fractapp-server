package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"fractapp-server/adaptors/polkascan"
	"fractapp-server/config"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"fractapp-server/scanner"
	"fractapp-server/types"

	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/go-pg/pg/v10"
)

var configPath = "config.json"

func init() {
	flag.StringVar(&configPath, "config", configPath, "config file")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	err := start(ctx, cancel)
	if err != nil {
		log.Fatal(err)
	}
}

func start(ctx context.Context, cancel context.CancelFunc) error {
	log.Info("Start scanner ...")
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

	pgDb := (*db.PgDB)(database)

	n, err := firebase.NewClient(ctx, "firebase.json", config.Firebase.ProjectId)
	if err != nil {
		return err
	}
	for k, url := range config.AdaptorUrls {
		network := types.ParseNetwork(k)
		adaptor := polkascan.NewAdaptor(url, network)
		bs := scanner.NewBlockScanner(pgDb, network.String(), network, n, adaptor)
		go func() {
			err = bs.Start()
			if err != nil {
				log.Errorf("%s scanner down: %s", network.String(), err)
			}
		}()
	}

	// await exit signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	return nil
}

package main

import (
	"context"
	"flag"
	"fractapp-server/config"
	"fractapp-server/db"
	"fractapp-server/push"
	"fractapp-server/scheduler"
	"os"
	"os/signal"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	log "github.com/sirupsen/logrus"
)

var (
	configPath = "config.json"
)

func init() {
	flag.StringVar(&configPath, "config", configPath, "config file")
	flag.Parse()
}

var database db.DB
var notificator push.Notificator

func main() {
	log.Info("Start scheduler...")

	ctx, cancel := context.WithCancel(context.Background())
	err := start(ctx, cancel)
	if err != nil {
		log.Fatal(err)
	}
}

func start(ctx context.Context, cancel context.CancelFunc) error {
	config, err := config.Parse(configPath)
	if err != nil {
		log.Fatalf("Invalid parse config: %s", err.Error())
	}

	notificator, err = push.NewClient(ctx, "firebase.json", config.Firebase.ProjectId)
	if err != nil {
		log.Fatalf("Invalid create notificator: %s", err.Error())
	}

	defer cancel()

	//TODO: add ctx with timeout
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DBConnectionString))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// Ping the primary
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	mongoDB, err := db.NewMongoDB(ctx, mongoClient)
	if err != nil {
		return err
	}

	go scheduler.Start(mongoDB, notificator, ctx)

	// await exit signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()

	return nil
}

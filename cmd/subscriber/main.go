package main

import (
	"context"
	"flag"
	"fractapp-server/config"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/subscriber"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	log "github.com/sirupsen/logrus"
)

var (
	host       = "127.0.0.1:9505"
	configPath = "config.json"
)

func init() {
	flag.StringVar(&host, "host", host, "host for server")
	flag.StringVar(&configPath, "config", configPath, "config file")
	flag.Parse()
}

func main() {
	log.Info("Start subscriber...")

	ctx, cancel := context.WithCancel(context.Background())
	config, err := config.Parse(configPath)
	if err != nil {
		log.Fatalf("Invalid parse config: %s", err.Error())
	}

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

	database, err := db.NewMongoDB(ctx, mongoClient)
	if err != nil {
		panic(err)
	}

	// create http server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	subController := subscriber.NewController(database)
	r.Group(func(r chi.Router) {
		r.Route(subController.MainRoute(), func(r chi.Router) {
			r.Post(subscriber.NotifyRoute, controller.Route(subController, subscriber.NotifyRoute))
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

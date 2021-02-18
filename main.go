package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"fractapp-server/config"
	"fractapp-server/controller"
	"fractapp-server/controller/auth"
	internalMiddleware "fractapp-server/controller/middleware"
	notificationController "fractapp-server/controller/notification"
	"fractapp-server/controller/profile"
	"fractapp-server/db"
	"fractapp-server/firebase"
	"fractapp-server/notification"
	"fractapp-server/scanner"
	"fractapp-server/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
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

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path + profile.AvatarDir); os.IsNotExist(err) {
		os.Mkdir(path+profile.AvatarDir, os.ModePerm)
	}

	// create http server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	pgDb := (*db.PgDB)(database)

	emailClient, err := notification.NewSMTPNotificator(config.Host, config.From.Address, config.From.Name, config.Password)
	if err != nil {
		return err
	}

	tokenAuth := jwtauth.New("HS256", []byte(config.Secret), nil)
	twilioApi := notification.NewTwilioNotificator(config.SMSService.FromNumber,
		config.SMSService.AccountSid, config.SMSService.AuthToken)

	nController := notificationController.NewController(pgDb)
	pController := profile.NewController(pgDb)
	authController := auth.NewController(
		pgDb,
		twilioApi,
		emailClient,
		tokenAuth,
	)

	authMiddleware := internalMiddleware.New(pgDb)

	r.Post(authController.MainRoute()+auth.SendCodeRoute, controller.Route(authController, auth.SendCodeRoute))
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.PubKeyAuth)
		r.Route(authController.MainRoute(), func(r chi.Router) {
			r.Post(auth.SignInRoute, controller.Route(authController, auth.SignInRoute))
		})
	})

	fs := http.FileServer("." + http.Dir(profile.AvatarDir))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(path + profile.AvatarDir + "/" + r.RequestURI); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			fs.ServeHTTP(w, r)
		}
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(authMiddleware.JWTAuth)

		r.Route(pController.MainRoute(), func(r chi.Router) {
			r.Get(profile.MyProfileRoute, controller.Route(pController, profile.MyProfileRoute))
			r.Get(profile.MyContactsRoute, controller.Route(pController, profile.MyContactsRoute))
			r.Get(profile.MyMatchContactsRoute, controller.Route(pController, profile.MyMatchContactsRoute))

			r.Post(profile.UpdateProfileRoute, controller.Route(pController, profile.UpdateProfileRoute))
			r.Post(profile.UploadAvatarRoute, controller.Route(pController, profile.UploadAvatarRoute))
			r.Post(profile.UploadContactsRoute, controller.Route(pController, profile.UploadContactsRoute))

		})
	})

	r.Group(func(r chi.Router) {
		r.Get(pController.MainRoute()+profile.UsernameRoute, controller.Route(pController, profile.UsernameRoute))
		r.Get(pController.MainRoute()+profile.UsernameRoute, controller.Route(pController, profile.UsernameRoute))
		r.Get(pController.MainRoute()+profile.SearchRoute, controller.Route(pController, profile.SearchRoute))
		r.Get(pController.MainRoute()+profile.InfoRoute, controller.Route(pController, profile.InfoRoute))

		r.Route(nController.MainRoute(), func(r chi.Router) {
			r.Post(notificationController.SubscribeRoute, controller.Route(nController, notificationController.SubscribeRoute))
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

	n, err := firebase.NewClient(ctx, config.Firebase.WithCredentialsFile, config.Firebase.ProjectId)
	if err != nil {
		return err
	}
	for k, url := range config.SubstrateUrls {
		network := types.ParseNetwork(k)
		es := scanner.NewEventScanner(url, pgDb, network.String(), network, n)
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

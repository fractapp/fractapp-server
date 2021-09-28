package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"fractapp-server/config"
	"fractapp-server/controller"
	"fractapp-server/controller/auth"
	"fractapp-server/controller/info"
	"fractapp-server/controller/message"
	internalMiddleware "fractapp-server/controller/middleware"
	"fractapp-server/controller/profile"
	"fractapp-server/controller/substrate"
	"fractapp-server/db"
	"fractapp-server/docs"
	"fractapp-server/notification"
	"fractapp-server/push"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
)

var host = "127.0.0.1:9544"
var configPath = "config.json"

// @contact.name Support
// @contact.email support@fractapp.com
// @license.name Apache 2.0
// @license.url https://github.com/fractapp/fractapp-server/blob/main/LICENSE
// @termsOfService https://fractapp.com/legal/tos.pdf

// @securityDefinitions.apikey AuthWithJWT
// @in header
// @name Authorization

// @securityDefinitions.apikey AuthWithPubKey-SignTimestamp
// @in header
// @name Sign-Timestamp

// @securityDefinitions.apikey AuthWithPubKey-Sign
// @in header
// @name Sign

// @securityDefinitions.apikey AuthWithPubKey-Auth-Key
// @in header
// @name Auth-Key

func init() {
	flag.StringVar(&host, "host", host, "host for server")
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
	defer cancel()
	log.Println("Setup api service")

	// parse config
	config, err := config.Parse(configPath)
	if err != nil {
		return errors.New(fmt.Sprint("Invalid parse config: ", err.Error()))
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

	mongoDB, err := db.NewMongoDB(ctx, mongoClient)
	if err != nil {
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

	emailClient, err := notification.NewSMTPNotificator(config.Host, config.From.Address, config.From.Name, config.Password)
	if err != nil {
		return err
	}

	tokenAuth := jwtauth.New("HS256", []byte(config.Secret), nil)
	twilioApi := notification.NewTwilioNotificator(config.SMSService.FromNumber,
		config.SMSService.AccountSid, config.SMSService.AuthToken)

	pController := profile.NewController(mongoDB, config.TransactionApi)
	substrateController := substrate.NewController(mongoDB, config.TransactionApi)

	authController := auth.NewController(
		mongoDB,
		twilioApi,
		emailClient,
		tokenAuth,
	)
	infoController := info.NewController(mongoDB)

	authMiddleware := internalMiddleware.New(mongoDB)

	notificator, err := push.NewClient(ctx, "firebase.json", config.Firebase.ProjectId)
	if err != nil {
		log.Fatalf("Invalid create notificator: %s", err.Error())
		return err
	}
	messageController := message.NewController(mongoDB, notificator)

	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Swagger Fractapp Server API"
	docs.SwaggerInfo.Description = "This is Fractapp server. Authorization flow described here: https://github.com/fractapp/fractapp-server/blob/main/AUTH.md"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/"
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(host+"/swagger/doc.json"),
	))

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.PubKeyAuth)
		r.Route(authController.MainRoute(), func(r chi.Router) {
			r.Post(auth.SignInRoute, controller.Route(authController, auth.SignInRoute))
		})
	})

	//TODO: will switch to another framework
	// Auth with JWT
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(authMiddleware.JWTAuth)

		r.Route(pController.MainRoute(), func(r chi.Router) {
			r.Get(profile.MyProfileRoute, controller.Route(pController, profile.MyProfileRoute))
			r.Get(profile.MyContactsRoute, controller.Route(pController, profile.MyContactsRoute))
			r.Get(profile.MyMatchContactsRoute, controller.Route(pController, profile.MyMatchContactsRoute))
			r.Post(profile.UpdateFirebaseTokenRoute, controller.Route(pController, profile.UpdateFirebaseTokenRoute))
			r.Post(profile.UpdateProfileRoute, controller.Route(pController, profile.UpdateProfileRoute))
			r.Post(profile.UploadAvatarRoute, controller.Route(pController, profile.UploadAvatarRoute))
			r.Post(profile.UploadContactsRoute, controller.Route(pController, profile.UploadContactsRoute))
		})

		r.Route(messageController.MainRoute(), func(r chi.Router) {
			r.Get(message.UnreadRoute, controller.Route(messageController, message.UnreadRoute))
			r.Post(message.SendRoute, controller.Route(messageController, message.SendRoute))
			r.Post(message.ReadRoute, controller.Route(messageController, message.ReadRoute))
		})
	})

	// Without Auth
	r.Group(func(r chi.Router) {
		r.Get(pController.MainRoute()+profile.AvatarRoute+"/*", controller.Route(pController, profile.AvatarRoute))
		r.Get(pController.MainRoute()+profile.UsernameRoute, controller.Route(pController, profile.UsernameRoute))
		r.Get(pController.MainRoute()+profile.SearchRoute, controller.Route(pController, profile.SearchRoute))
		r.Get(pController.MainRoute()+profile.UserInfoRoute, controller.Route(pController, profile.UserInfoRoute))
		r.Get(pController.MainRoute()+profile.TransactionStatusRoute, controller.Route(pController, profile.TransactionStatusRoute))
		r.Get(pController.MainRoute()+profile.TransactionsRoute, controller.Route(pController, profile.TransactionsRoute))

		r.Get(infoController.MainRoute()+info.TotalRoute, controller.Route(infoController, info.TotalRoute))

		r.Post(authController.MainRoute()+auth.SendCodeRoute, controller.Route(authController, auth.SendCodeRoute))

		r.Get(substrateController.MainRoute()+substrate.FeeRoute, controller.Route(substrateController, substrate.FeeRoute))
		r.Get(substrateController.MainRoute()+substrate.BaseRoute, controller.Route(substrateController, substrate.BaseRoute))
		r.Get(substrateController.MainRoute()+substrate.TxBaseRoute, controller.Route(substrateController, substrate.TxBaseRoute))
		r.Post(substrateController.MainRoute()+substrate.BroadcastRoute, controller.Route(substrateController, substrate.BroadcastRoute))
		r.Get(substrateController.MainRoute()+substrate.BalanceRoute, controller.Route(substrateController, substrate.BalanceRoute))
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

	return nil
}

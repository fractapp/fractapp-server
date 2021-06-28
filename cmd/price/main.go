package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"fractapp-server/config"
	"fractapp-server/db"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	log "github.com/sirupsen/logrus"
)

var (
	currency   = "DOT"
	configPath = "config.json"
	startTime  = int64(1597622400000) // Mon Aug 17 2020 00:00:00 GMT+0000

	MaxRqCountErr  = errors.New("request limit reached")
	RqHttpLimitErr = errors.New("request limit reached (http)")
	IpBannedErr    = errors.New("ip is banned")

	binanceApi = ""
)

const (
	interval   = 5 // history price interval (minutes)
	limit      = 1000
	maxRqCount = 1200
)

func init() {
	flag.StringVar(&currency, "currency", currency, "currency")
	flag.StringVar(&configPath, "config", configPath, "config path")
	flag.Int64Var(&startTime, "start", startTime, "start time for scanning")
	flag.Parse()
}

func main() {
	log.Info("Start price cache ...")

	ctx, cancel := context.WithCancel(context.Background())
	config, err := config.Parse(configPath)
	if err != nil {
		log.Fatalf("Invalid parse config: %s", err.Error())
	}

	binanceApi = config.BinanceApi

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
		panic(err)
	}

	go func() {
		for {
			err := startScanForCurrency(mongoDB, ctx)
			if err != nil {
				log.Errorf("invalid start scan: %s", err)
				continue
			}
			log.Info("Wait new price")
			time.Sleep(time.Minute) //TODO (calculate wait for time)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
}

func startScanForCurrency(database db.DB, ctx context.Context) error {
	iterator := (time.Minute * interval * limit).Milliseconds()
	lastPrice, err := database.LastPriceByCurrency(currency)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	if lastPrice != nil {
		startTime = lastPrice.Timestamp
	}

	for start := startTime; start < time.Now().Unix()*1000; start += iterator {
		isWritten := false
		for !isWritten {
			err := scan(start, start+iterator, database, ctx)

			switch err {
			case RqHttpLimitErr:
				log.Error(err)
				log.Info("Wait 1 minute")
				time.Sleep(1 * time.Minute)
			case IpBannedErr:
				log.Error(err)
				log.Info("Wait 5 minute")
				time.Sleep(5 * time.Minute)
			case MaxRqCountErr:
				log.Info("Wait 1 minute")
				time.Sleep(1 * time.Minute)
				fallthrough
			case nil:
				isWritten = true
			default:
				log.Error(err)
				continue
			}
		}
	}

	return nil
}

func scan(startTime int64, endTime int64, database db.DB, ctx context.Context) error {
	log.Infof("scan start time: %d", startTime/1000)
	log.Infof("scan end time: %d", endTime/1000)

	resp, err := http.Get(fmt.Sprintf(
		"https://%s/api/v3/klines?symbol=%sUSDT&startTime=%d&endTime=%d&limit=%d&interval=%dm",
		binanceApi,
		currency, startTime, endTime, limit, interval),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return RqHttpLimitErr
	} else if resp.StatusCode == 418 {
		return IpBannedErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	prices := make([][]interface{}, 0, 0)
	err = json.Unmarshal(body, &prices)
	if err != nil {
		return err
	}

	weight, _ := strconv.Atoi(resp.Header.Get("x-mbx-used-weight"))
	log.Infof("Weight: %d", weight)
	if weight > maxRqCount {
		return MaxRqCountErr
	}

	pricesLen := len(prices)
	if pricesLen <= 0 {
		return nil
	}

	log.Infof("Start insert to db ...")
	if pricesLen >= 100 {
		divider := pricesLen / 10
		chArray := make([]chan bool, 0)

		for i := 0; i < 8; i++ {
			chArray = append(chArray, write(prices, divider*i, divider*(i+1), database))
		}
		chArray = append(chArray, write(prices, divider*9, pricesLen, database))

		for _, v := range chArray {
			<-v
		}
	} else {
		write(prices, 0, pricesLen, database)
	}

	log.Infof("End insert to db")
	log.Info("-------------------------------------------")

	return nil
}

func write(prices [][]interface{}, startIndex int, endIndex int, database db.DB) chan bool {
	var dbPrices []interface{}

	now := time.Now().Unix() * 1000
	for _, v := range prices[startIndex:endIndex] {
		timestamp := int64(v[6].(float64))
		diff := timestamp - now
		if diff > 0 {
			continue
		}

		price, _ := strconv.ParseFloat(v[4].(string), 32)
		dbPrices = append(dbPrices, &db.Price{
			Timestamp: timestamp,
			Currency:  currency,
			Price:     float32(price),
		})
	}
	chResult := make(chan bool)
	go func() {
		if len(dbPrices) == 0 {
			chResult <- true
			return
		}
		err := database.InsertMany(dbPrices)
		if err != nil {
			log.Errorf("Insert price to db (%d:%d): %s", startIndex, endIndex, err)
		}
		chResult <- true
	}()

	return chResult
}

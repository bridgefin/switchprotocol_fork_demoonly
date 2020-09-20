package main

import (
	"flag"
	"sync"
	"time"

	scrapers "github.com/bridgefinance.net/bridgefinance/internal/pkg/exchange-scrapers"
	"github.com/bridgefinance.net/bridgefinance/pkg/BFR"
	"github.com/bridgefinance.net/bridgefinance/pkg/BFR/helpers/configCollectors"
	"github.com/bridgefinance.net/bridgefinance/pkg/BFR/helpers/kafkaHelper"
	models "github.com/bridgefinance.net/bridgefinance/pkg/model"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

const (
	watchdogDelay = 60.0 * 20
)

func handleTrades(c chan *BFR.Trade, wg *sync.WaitGroup, w *kafka.Writer) {
	lastTradeTime := time.Now()
	t := time.NewTicker(watchdogDelay * time.Second)
	for {
		select {
		case <-t.C:
			duration := time.Since(lastTradeTime)
			if duration.Seconds() > watchdogDelay {
				log.Error(duration)
				panic("frozen? ")
			}
		case t, ok := <-c:
			if !ok {
				wg.Done()
				log.Error("handleTrades")
				return
			}
			lastTradeTime = time.Now()
			kafkaHelper.WriteMessage(w, t)
		}
	}
}

var (
	exchange         = flag.String("exchange", "", "which exchange")
	onePairPerSymbol = flag.Bool("onePairPerSymbol", false, "one Pair max Per Symbol ?")
)

func init() {
	flag.Parse()
	if *exchange == "" {
		flag.Usage()
		log.Println(BFR.Exchanges())
		log.Fatal("exchange is required")
	}
}

// main manages all PairScrapers and handles incoming trade information
func main() {

	ds, err := models.NewRedisDataStore()
	if err != nil {
		log.Errorln("NewDataStore:", err)
	} else {

	}
	pairsExchange, err := ds.GetAvailablePairsForExchange(*exchange)

	if err != nil || len(pairsExchange) == 0 {
		log.Error("error on GetAvailablePairsForExchange", err)
		cc := configCollectors.NewConfigCollectors(*exchange)
		pairsExchange = cc.AllPairs()
	}

	configApi, err := BFR.GetConfig(*exchange)
	if err != nil {
		log.Warning("no config for exchange's api ", err)
	}
	es := scrapers.NewAPIScraper(*exchange, configApi.ApiKey, configApi.SecretKey)

	w := kafkaHelper.NewWriter(kafkaHelper.TopicTrades)
	defer w.Close()

	wg := sync.WaitGroup{}

	pairs := make(map[string]string)

	for _, configPair := range pairsExchange {
		dontAddPair := false
		if *onePairPerSymbol {
			_, dontAddPair = pairs[configPair.Symbol]
			pairs[configPair.Symbol] = configPair.Symbol
		}
		if dontAddPair {
			log.Println("Skipping pair:", configPair.Symbol, configPair.ForeignName, "on exchange", *exchange)
		} else {
			log.Println("Adding pair:", configPair.Symbol, configPair.ForeignName, "on exchange", *exchange)
			_, err := es.ScrapePair(BFR.Pair{
				Symbol:      configPair.Symbol,
				ForeignName: configPair.ForeignName})
			if err != nil {
				log.Println(err)
			} else {
				wg.Add(1)
			}
		}
		defer wg.Wait()
	}
	go handleTrades(es.Channel(), &wg, w)
}
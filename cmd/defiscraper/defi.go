package main

import (
	"flag"
	"sync"

	"github.com/bridgefinance.net/bridgefinance/pkg/BFR"

	defiscraper "github.com/bridgefinance.net/bridgefinance/internal/pkg/defiscrapers"
	models "github.com/bridgefinance.net/bridgefinance/pkg/model"
	log "github.com/sirupsen/logrus"
)

// handleDefiInterestRate delegates rate information to Kafka
func handleDefiInterestRate(c chan *BFR.DefiRate, wg *sync.WaitGroup, ds models.Datastore) {
	defer wg.Done()
	// Pull from channel as long as not empty
	for {
		t, ok := <-c
		if !ok {
			log.Error("error")
			return
		}
		ds.SetDefiRateInflux(t)
	}
}

// handleDefiState delegates rate information to Kafka
func handleDefiState(c chan *BFR.DefiProtocolState, wg *sync.WaitGroup, ds models.Datastore) {
	defer wg.Done()
	// Pull from channel as long as not empty
	for {
		t, ok := <-c
		if !ok {
			log.Error("error")
			return
		}
		ds.SetDefiStateInflux(t)
	}
}

func main() {
	rateType := flag.String("type", "DYDX", "Type of Defi rate")
	flag.Parse()

	wg := sync.WaitGroup{}
	ds, err := models.NewDataStore()

	if err != nil {
		log.Errorln("NewDataStore:", err)
	} else {

		sRate := defiscraper.SpawnDefiScraper(ds, *rateType)
		defer sRate.Close()

		// Send rates to the database while the scraper scrapes
		wg.Add(2)
		go handleDefiInterestRate(sRate.RateChannel(), &wg, ds)
		go handleDefiState(sRate.StateChannel(), &wg, ds)

		defer wg.Wait()
	}
}

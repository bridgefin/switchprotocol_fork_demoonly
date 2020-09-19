package main

import (
	"github.com/bridgefinance-net/bridgefinance/internal/pkg/exchange-scrapers"
	_ "github.com/bridgefinance-net/bridgefinance/pkg/BFR"
	_ "github.com/bridgefinance-net/bridgefinance/pkg/model"
	"sync"
	log "github.com/sirupsen/logrus"
)

// TODO: Read the instruments from DB (formerly: deribitOptionsMetaFilename)
// pairs contains all pairs currently supported by the BFR scrapers
var (
	instruments = []string{
		"BTC-27MAR20-9000-C",
		"BTC-24APR20-9000-C",
		"BTC-27MAR20-9000-P",
		"BTC-24APR20-9000-P",
	}
)

// main manages all Scrapers and handles incoming trade information
func main() {
	wg := sync.WaitGroup{}
	allScrapers := scrapers.NewAllDeribitOptionsScrapers(&wg, instruments, "rHQ8rYAo", "UmX8Ea0FelZzvT0nB44ZUWdRu6jyBYUMZDqE_gtQ2us")
	go allScrapers.GetMetas()
	go allScrapers.RefreshMetas("BTC")
	allScrapers.ScrapeMarkets()

	log.Info(wg)
	defer wg.Wait()
}

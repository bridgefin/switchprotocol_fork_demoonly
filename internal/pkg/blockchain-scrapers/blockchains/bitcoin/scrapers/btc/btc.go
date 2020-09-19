package main

import (
	"github.com/bridgefinance-net/bridgefinance/internal/pkg/blockchain-scrapers/scrapers"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
)

const (
	SERVER_HOST = "bitcoin"
	SERVER_PORT = 8332
	USER        = "mysecretrpcbfruser"
	PASSWD      = "mysecretrpcbfrpassword"
	SYMBOL      = "BTC"
	TIP_TIME    = 60 * 60 * 2
)

func main() {
	config := bfr.GetConfigApi()
	if config == nil {
		panic("Couldnt load config")
	}
	client := bfr.NewClient(config)
	if client == nil {
		panic("Couldnt load client")
	}
	scraper := blockchainscrapers.NewScraper(client, SYMBOL, SERVER_HOST, SERVER_PORT, USER, PASSWD, TIP_TIME)
	if scraper == nil {
		panic("Couldnt load scraper")
	}
	scraper.Run()
}
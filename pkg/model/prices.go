package models

import (
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	log "github.com/sirupsen/logrus"
	"time"
)

func (db *DB) SetPriceZSET(symbol string, exchange string, price float64, t time.Time) error {
	db.SaveFilterInflux(bfr.FilterKing, symbol, exchange, price, t)
	key := getKeyFilterZSET(getKey(bfr.FilterKing, symbol, exchange))
	log.Debug("SetPriceZSET ", key)
	return db.setZSETValue(key, price, time.Now().Unix(), BiggestWindow)
}

func (db *DB) GetPrice(symbol string, exchange string) (float64, error) {
	key := getKeyFilterSymbolAndExchangeZSET(bfr.FilterKing, symbol, exchange)
	v, _, err := db.getZSETLastValue(key)
	return v, err
}

func (db *DB) GetPriceYesterday(symbol string, exchange string) (float64, error) {
	return db.getZSETValue(getKeyFilterZSET(getKey(bfr.FilterKing, symbol, exchange)), time.Now().Unix()-WindowYesterday)
}

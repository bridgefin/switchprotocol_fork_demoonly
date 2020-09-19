package models

import (
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func getKeySupply(value string) string {
	return "bfr_supply_" + value
}

func (db *DB) SymbolsWithASupply() ([]string, error) {
	result := []string{}
	var cursor uint64
	key := getKeySupply("")
	for {
		var keys []string
		var err error
		keys, cursor, err = db.redisClient.Scan(cursor, key+"*", 10).Result()
		if err != nil {
			log.Error("SymbolsWithASupply err", err)
			return result, err
		}
		for _, value := range keys {
			result = append(result, strings.Replace(value, key, "", 1))
		}
		if cursor == 0 {
			log.Debugf("SymbolsWithASupply %v returns %v", key, result)
			return result, nil
		}
	}
}

func (a *DB) GetLatestSupply(symbol string) (*bfr.Supply, error) {
	val, err := a.GetSupply(symbol, time.Time{}, time.Time{})
	if err != nil {
		log.Error(err)
		return &bfr.Supply{}, err
	}
	return &val[0], err
}

func (a *DB) GetSupply(symbol string, starttime, endtime time.Time) ([]bfr.Supply, error) {
	switch symbol {
	case "MIOTA":
		retArray := []bfr.Supply{}
		s := bfr.Supply{
			Symbol:            "MIOTA",
			CirculatingSupply: 2779530283.0,
			Name:              helpers.NameForSymbol("MIOTA"),
			Time:              time.Now(),
			Source:            bfr.bridgefinance,
		}
		retArray = append(retArray, s)
		return retArray, nil
	default:
		value, err := a.GetSupplyInflux(symbol, starttime, endtime)
		if err != nil {
			log.Errorf("Error: %v on GetSupply %v\n", err, symbol)
			return []bfr.Supply{}, err
		}
		return value, err
	}
}

func (a *DB) SetSupply(supply *bfr.Supply) error {
	key := getKeySupply(supply.Symbol)
	log.Debug("setting ", key, supply)
	err := a.redisClient.Set(key, supply, 0).Err()
	if err != nil {
		log.Errorf("Error: %v on SetSupply (redis) %v\n", err, supply.Symbol)
	}
	err = a.SaveSupplyInflux(supply)
	if err != nil {
		log.Errorf("Error: %v on SetSupply (influx) %v\n", err, supply.Symbol)
	}
	return err
}

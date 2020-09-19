package models

import (
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	log "github.com/sirupsen/logrus"
	"strings"
)

// exchange = "" for all exchanges
func (db *DB) GetPairs(exchange string) ([]bfr.Pair, error) {
	var result []bfr.Pair
	var cursor uint64
	key := "bfr_" + bfr.FilterKing + "_"
	for {
		var keys []string
		var err error
		keys, cursor, err = db.redisClient.Scan(cursor, key+"*", 10).Result()
		if err != nil {
			log.Error("GetPairs err", err)
			return result, err
		}
		for _, value := range keys {
			filteredKey := strings.Replace(strings.Replace(value, key, "", 1), "_ZSET", "", 1)
			s := strings.Split(strings.Replace(filteredKey, key, "", 1), "_")
			if len(s) == 2 {
				result = append(result, bfr.Pair{
					Symbol:   s[0],
					Exchange: s[1],
				})
			}
		}
		if cursor == 0 {
			log.Debugf("GetPairs %v returns %v", key, result)
			return result, nil
		}
	}
}

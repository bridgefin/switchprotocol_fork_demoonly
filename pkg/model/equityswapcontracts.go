package models

import (
  log "github.com/sirupsen/logrus"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	"errors"
	"time"
)

func (db *DB) SetOptionMeta(optionMeta *bfr.OptionMeta) error {
	if db.redisClient == nil {
		return errors.New("Datastore has no redis client.")
	}
	key := "bfr_optionMeta_" + optionMeta.BaseCurrency
	log.Debug("setting ", key, optionMeta)
	err := db.redisClient.SAdd(key, optionMeta).Err()
	if err != nil {
		log.Printf("Error: %v on SetOptionMeta %v\n", err, key)
	}
	return err
}

func (db *DB) RemoveExpiredOptionMeta(baseCurrency string) error {
	if db.redisClient == nil {
		return errors.New("Datastore has no redis client.")
	}
	optionsMeta, err := db.GetOptionMeta(baseCurrency)
	if err != nil {
		return err
	}
	key := "bfr_optionMeta_" + baseCurrency
	for _, optionMeta := range optionsMeta {
		if optionMeta.ExpirationTime.Before(time.Now()) {
			err = db.redisClient.SRem(key, optionMeta).Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) GetOptionMeta(baseCurrency string) ([]bfr.OptionMeta, error) {
	var result []bfr.OptionMeta
	if db.redisClient == nil {
		return result, errors.New("Datastore has no redis client.")
	}
	key := "bfr_optionMeta_" + baseCurrency
	resultStrings, err := db.redisClient.SMembers(key).Result()

	if err != nil {
		log.Error("GetOptionMeta: ", err)
	}

	for _, v := range resultStrings {
		currentOM := bfr.OptionMeta{}
		err = currentOM.UnmarshalBinary([]byte(v))
		if err != nil {
			log.Error("GetOptionMeta: ", err)
		}
		result = append(result, currentOM)
	}
	return result, err
}

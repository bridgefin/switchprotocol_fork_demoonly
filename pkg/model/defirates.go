package models

import (
	"encoding/json"

	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
)

// SetDefiProtocol writes @protocol into redis
func (db *DB) SetDefiProtocol(protocol bfr.DefiProtocol) error {
	keyProtocol := "bfr_DefiProtocol_" + protocol.Name
	mProtocol, err := json.Marshal(protocol)
	if err != nil {
		return err
	}
	err = db.redisClient.Set(keyProtocol, mProtocol, TimeOutRedis).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetDefiProtocol returns the die protocol struct by name
func (db *DB) GetDefiProtocol(name string) (bfr.DefiProtocol, error) {
	protocol := bfr.DefiProtocol{}
	key := "bfr_DefiProtocol_" + name
	err := db.redisClient.Get(key).Scan(&protocol)
	if err != nil {
		return protocol, err
	}
	return protocol, nil
}

// GetDefiProtocols returns a slice of all available DeFi protocols
func (db *DB) GetDefiProtocols() ([]bfr.DefiProtocol, error) {
	allProtocols := []bfr.DefiProtocol{}
	pattern := "bfr_DefiProtocol_*"
	allKeys := db.redisClient.Keys(pattern).Val()
	for _, key := range allKeys {
		protocol := bfr.DefiProtocol{}
		err := db.redisClient.Get(key).Scan(&protocol)
		if err != nil {
			return []bfr.DefiProtocol{}, err
		}
		allProtocols = append(allProtocols, protocol)
	}
	return allProtocols, nil
}

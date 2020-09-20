package main

import (
	"context"
	"github.com/bridgefinance-net/bridgefinance/internal/pkg/tradesBlockService"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr/helpers/configCollectors"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr/helpers/kafkaHelper"
	"github.com/bridgefinance-net/bridgefinance/pkg/model"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"sync"
)

func handleBlocks(blockMaker *tradesBlockService.TradesBlockService, wg *sync.WaitGroup, w *kafka.Writer) {
	for {
		t, ok := <-blockMaker.Channel()
		if !ok {
			log.Printf("handleBlocks: finishing channel")
			wg.Done()
			return
		}
		err := kafkaHelper.WriteMessage(w, t)
		if err != nil {
			log.Errorln("handleBlocks", err)
		}
	}
}

func main() {

	cc := configCollectors.NewConfigCollectorsIfExists("vip")
	if cc == nil {
		log.Warning("no vip.json: accepting all trades.")
	}

	w := kafkaHelper.NewSyncWriter(kafkaHelper.TopicTradesBlock)
	defer w.Close()

	r := kafkaHelper.NewReaderNextMessage(kafkaHelper.TopicTrades)
	defer r.Close()

	s, err := models.NewDataStore()
	if err != nil {
		log.Errorln("NewDataStore", err)
	}

	tradesBlockService := tradesBlockService.NewTradesBlockService(s, bfr.BlockSizeSeconds)

	wg := sync.WaitGroup{}
	go handleBlocks(tradesBlockService, &wg, w)

	log.Printf("starting...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf(err.Error())
		} else {
			var t bfr.Trade
			err := t.UnmarshalBinary(m.Value)
			if err == nil {
				if cc != nil {
					if cc.IsSymbolInConfig(t.Symbol) == false {
						continue
					}
				}
				tradesBlockService.ProcessTrade(&t)
			} else {
				log.Printf("ignored message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
			}
		}
	}
}
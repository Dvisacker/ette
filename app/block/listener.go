package block

import (
	"context"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(client *ethclient.Client, _db *gorm.DB, _lock *sync.Mutex, _synced *d.SyncState, redisClient *redis.Client) {
	headerChan := make(chan *types.Header)

	subs, err := client.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatalf("[!] Failed to subscribe to block headers : %s\n", err.Error())
	}

	// Scheduling unsubscribe, to be executed when end of this block is reached
	defer subs.Unsubscribe()

	// Flag to check for whether this is first time block header being received
	// or not
	//
	// If yes, we'll start syncer to fetch all block starting from 0 to this block
	first := true

	for {
		select {
		case err := <-subs.Err():
			log.Printf("[!] Block header subscription failed in mid : %s\n", err.Error())
			break
		case header := <-headerChan:
			if first {
				// Starting syncer in another thread, where it'll keep fetching
				// blocks starting from genesis to this block
				go SyncToLatestBlock(client, _db, 0, header.Number.Uint64(), _lock, _synced)
				// Making sure on when next latest block header is received, it'll not
				// start another syncer
				first = false
			}

			if err := redisClient.Publish(context.Background(), "block", &d.Block{
				Hash:       header.Hash().Hex(),
				Number:     header.Number.Uint64(),
				Time:       header.Time,
				ParentHash: header.ParentHash.Hex(),
				Difficulty: header.Difficulty.String(),
				GasUsed:    header.GasUsed,
				GasLimit:   header.GasLimit,
				Nonce:      header.Nonce.Uint64(),
			}).Err(); err != nil {

				log.Printf("[!] Failed to publish block %d in channel : %s\n", header.Number.Uint64(), err.Error())

			}

			go fetchBlockByHash(client, header.Hash(), _db, redisClient)
		}
	}
}

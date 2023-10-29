package batchprocessor

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const DEFAULT_RATE_LIMIT_PERIOD_TIME = 1
const DEFAULT_RATE_LIMIT_UNIT = time.Second

var collectedData []interface{}
var batchBlocked = false
var wg sync.WaitGroup
var mu1 sync.Mutex
var mu2 sync.Mutex
var mu3 sync.Mutex
var batchResult []interface{}
var batchCondition = sync.NewCond(&mu3)
var counter = 0

type BatchingQueueingProcessor interface {
	Process(data interface{}, process func(data []interface{}) []interface{}) []interface{}
}

type batchingQueuingProcessor struct {
	cfg Config
}

func New() BatchingQueueingProcessor {
	return batchingQueuingProcessor{
		cfg: Config{
			RateLimitTimePeriod: DEFAULT_RATE_LIMIT_PERIOD_TIME,
			RateLimitUnit:       DEFAULT_RATE_LIMIT_UNIT,
			logger:              log.Default(),
		},
	}
}

func NewWithConfig(cfg Config) BatchingQueueingProcessor {
	return batchingQueuingProcessor{
		cfg: cfg,
	}
}

func (b batchingQueuingProcessor) Process(data interface{}, process func(data []interface{}) []interface{}) []interface{} {
	// Stop before adding to the batch if a request is in progress
	batchCondition.L.Lock()
	for batchBlocked {
		fmt.Printf("ID %+v Waiting for batch request to finish\n", data)
		batchCondition.Wait()
	}
	batchCondition.L.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Gather all data for the batch request
		mu1.Lock()
		fmt.Printf("ID %+v Adding data for batch request\n", data)
		collectedData = append(collectedData, data)
		counter++
		mu1.Unlock()

		if mu2.TryLock() {
			// Check if the next request is within the waiting period.
			fmt.Printf("ID %+v Waiting for more requests\n", data)
			select {
			case <-time.After(b.cfg.RateLimitTimePeriod * b.cfg.RateLimitUnit):
				batchBlocked = true
				// Send the collected data to the processing function.
				batchResult = process(collectedData)
			}
			// Clear the collected data.
			collectedData = nil
			mu2.Unlock()
		}
	}()

	// Wait for all goroutines to finish.
	wg.Wait()

	mu1.Lock()
	counter--
	// The last request to be served cleans the variables and wakes the threads waiting for the next batch
	if counter == 0 {
		fmt.Printf("ID %+v last routine resets the variables and awakes the waiting threads\n\n", data)
		defer func() {
			batchBlocked = false
			batchCondition.Broadcast()
		}()
	}
	mu1.Unlock()

	return batchResult
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const RATE_LIMIT_PERIOD_TIME = 1
const RATE_LIMIT_UNIT = time.Second

var collectedData []string
var batchBlocked = false
var wg sync.WaitGroup
var mu1 sync.Mutex
var mu2 sync.Mutex
var mu3 sync.Mutex
var batchResult string
var batchCondition = sync.NewCond(&mu3)
var counter = 0

func main() {
	http.HandleFunc("/process", processHandler)
	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ID := parseID(r)
	result := fmt.Sprintf("|ID: %s|", ID)

	// Stop before adding to the batch if a request is in progress
	batchCondition.L.Lock()
	for batchBlocked {
		fmt.Printf("ID %+v Waiting for batch request to finish\n", ID)
		batchCondition.Wait()
	}
	batchCondition.L.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Gather all data for the batch request
		mu1.Lock()
		fmt.Printf("ID %+v Adding data for batch request\n", ID)
		collectedData = append(collectedData, result)
		counter++
		mu1.Unlock()

		if mu2.TryLock() {
			// Check if the next request is within the waiting period.
			fmt.Printf("ID %+v Waiting for more requests\n", ID)
			select {
			case <-time.After(RATE_LIMIT_PERIOD_TIME * RATE_LIMIT_UNIT):
				batchBlocked = true
				// Send the collected data to the processing function.
				batchResult = processData(collectedData, ID)
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
		fmt.Printf("ID %+v last routine resets the variables and awakes the waiting threads\n\n", ID)
		defer func() {
			batchBlocked = false
			batchCondition.Broadcast()
		}()
	}
	mu1.Unlock()

	// Measure time until response
	elapsed := time.Since(start)
	log.Printf("ID: %+v Took %s", ID, elapsed)

	// Respond to the original request.
	fmt.Fprintf(w, batchResult)
}

func processData(data []string, ID string) string {
	s := fmt.Sprintf("ID: %+v Processing data: %v\n", ID, data)
	fmt.Printf(s)
	// Simulate some work being done with random time
	sleepRandomTime()
	return s
}

func parseID(r *http.Request) string {
	requestData := r.URL
	u, err := url.Parse(requestData.String())
	if err != nil {
		panic(err)
	}
	m, _ := url.ParseQuery(u.RawQuery)
	ID := m["ID"][0]

	return ID
}

func sleepRandomTime() {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(1000)
	minSleep := 100 * time.Millisecond
	time.Sleep(minSleep + time.Duration(n)*time.Millisecond)
}

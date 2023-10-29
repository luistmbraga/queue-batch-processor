package main

import (
	internal "batch-request-service/pkg/batchprocessor"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var processor internal.BatchingQueueingProcessor

func main() {
	processor = internal.New()
	http.HandleFunc("/process", processHandler)
	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ID := parseID(r)
	data := fmt.Sprintf("|ID: %s|", ID)

	batchResult := processor.Process(data, processData)

	// Measure time until response
	elapsed := time.Since(start)
	log.Printf("ID: %+v Took %s", ID, elapsed)

	// Respond to the original request.
	fmt.Fprintf(w, fmt.Sprintf("%+v", batchResult))
}

func processData(data []interface{}) []interface{} {
	// Simulate some work being done with random time
	sleepRandomTime()
	return data
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

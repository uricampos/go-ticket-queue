package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var MAX_REQUESTS = 2000

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("no .env file found, relying on real env vars")
	}

	userID := os.Getenv("TEST_USER_ID")
	if userID == "" {
		panic("TEST_USER_ID not set")
	}

	postUrl := "http://localhost:5001/orders"

	body := fmt.Appendf(nil, `{"user_id": "%s", "total_price": 100}`, userID)

	var wg sync.WaitGroup

	statusCodes := make([]int, MAX_REQUESTS)
	errs := make([]error, MAX_REQUESTS)

	start := time.Now()

	idempotencyKey := "abcde123"

	for i := range MAX_REQUESTS {
		wg.Add(1)
		go makeRequest(postUrl, body, &wg, i, statusCodes, errs, idempotencyKey)
	}

	wg.Wait()
	elapsed := time.Since(start)

	total201 := 0
	totalErrs := 0

	for t := range statusCodes {
		if statusCodes[t] == 201 {
			total201 += 1
		}
	}

	for e := range errs {
		if errs[e] != nil {
			totalErrs += 1
		}
	}

	fmt.Printf("Total 201: %d\n", total201)
	fmt.Printf("Total errors: %d\n", totalErrs)
	fmt.Printf("Running time: %v\n", elapsed)
}

func makeRequest(postUrl string, body []byte, wg *sync.WaitGroup, i int, statusCodes []int, errs []error, idempotencyKey string) {
	defer wg.Done()
	r, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(body))

	if err != nil {
		errs[i] = err
		return
	}

	// r.Header.Add("Idempotency-Key", uuid.NewString())
	r.Header.Add("Idempotency-Key", idempotencyKey)

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		errs[i] = err
		return
	}

	defer resp.Body.Close()
	statusCodes[i] = resp.StatusCode
}

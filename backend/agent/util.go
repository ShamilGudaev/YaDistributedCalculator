package agent

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	rand "math/rand"
	"net/http"
	"time"
)

func printIfHttpReqFailed(err error, agentID *string, path string, d time.Duration) bool {
	if err != nil {
		fmt.Printf("%s: error making http request (%s): %s\n", *agentID, path, err)
		time.Sleep(d)
		return true
	}
	return false
}

func panicIfBadRequest(res *http.Response, agentID *string, path string) {
	if res.StatusCode == http.StatusBadRequest {
		panic(fmt.Sprintf("%s: Bad Request (%s)", *agentID, path))
	}
}

func printIfInternalServerError(res *http.Response, agentID *string, path string, d time.Duration) bool {
	if res.StatusCode == http.StatusInternalServerError {
		fmt.Printf("%s: Internal Server Error (%s)\n", *agentID, path)
		time.Sleep(d)
		return true
	}

	return false
}

func printIfResponseIsInvalid(e error, agentID *string, path string, d time.Duration) bool {
	if e != nil {
		fmt.Printf("%s: Orchestrator response is invalid (%s)\n", *agentID, path)
		time.Sleep(d)
		return true
	}

	return false
}

func panicIfNotOk(res *http.Response, agentID *string, path string) {
	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("%s: Unexpected status code (%s): %s", *agentID, path, res.Status))
	}
}

func LazyInterval(out chan bool, d time.Duration) {
	for {
		time.Sleep(d)
		out <- true
	}
}

func LazyIntervalRange(out chan bool, from time.Duration, to time.Duration) {
	seed := new(int64)
	err := binary.Read(crand.Reader, binary.BigEndian, seed)
	if err != nil {
		log.Fatal(err)
	}
	rng := rand.New(rand.NewSource(*seed))
	diff := (to - from).Nanoseconds()
	for {
		time.Sleep(from + time.Duration(rng.Int63n(diff))*time.Nanosecond)
		out <- true
	}
}

package agent

import (
	crand "crypto/rand"
	"encoding/binary"
	"log"
	rand "math/rand"
	"time"
)

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

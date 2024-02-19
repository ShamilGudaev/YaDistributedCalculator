package main

import (
	"backend/agent"
	"os"
	"strconv"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var limit = 4

func main() {
	res, err := strconv.Atoi(os.Getenv("WORKERS_LIMIT"))

	if err != nil {
		panic(err)
	}

	if res <= 0 {
		panic("WORKERS_LIMIT must be greater than 1")
	}

	limit = res

	for {
		startAgent()
	}
}

var workers map[uint64]*agent.Worker = make(map[uint64]*agent.Worker)

func startAgent() {
	agentId, err := gonanoid.New()
	if err != nil {
		panic(err)
	}

	defer stopWorkers()

	aliveInterval := make(chan bool, 1)
	go agent.LazyInterval(aliveInterval, 5*time.Second)

	requestInterval := make(chan bool, 1)
	go agent.LazyIntervalRange(requestInterval, 2*time.Second, 3*time.Second)

	resultChannel := make(chan agent.Result, limit)

	for {
		select {
		case result := <-resultChannel:
			if !agent.SubmitResult(agentId, result.ExpressionID, result.Result) {
				return
			}
			delete(workers, result.ExpressionID)

		case <-aliveInterval:
			if !agent.SendAliveMessage(agentId) {
				return
			}
		case <-requestInterval:
			if len(workers) >= limit {
				continue
			}

			res := agent.RequestExpression(agentId)
			if res.IsDeleted {
				return
			}

			data := res.Data

			if data == nil {
				continue
			}

			worker := &agent.Worker{
				ExpressionID: data.ExpressionID,
				StopChannel:  make(chan bool, 1),
			}

			workers[worker.ExpressionID] = worker
			go worker.Start(data, resultChannel)
		}
	}

}

func stopWorkers() {
	for _, worker := range workers {
		worker.StopChannel <- true
	}

	workers = make(map[uint64]*agent.Worker)
}

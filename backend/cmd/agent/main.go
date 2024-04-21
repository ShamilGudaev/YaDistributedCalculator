package main

import (
	"backend/agent"
	"log"
	"os"
	"strconv"
	"time"

	pb "backend/proto"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var limit = 4

func main() {
	limitEnv := os.Getenv("WORKERS_LIMIT")

	if limitEnv != "" {
		res, err := strconv.Atoi(limitEnv)

		if err != nil {
			panic(err)
		}

		if res <= 0 {
			panic("WORKERS_LIMIT should be >= 1")
		}

		limit = res
	}

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

	keepAliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.Dial("orchestrator:1324", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepAliveParams))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrchestratorClient(conn)

	aliveInterval := make(chan bool, 1)
	go agent.LazyInterval(aliveInterval, 5*time.Second)

	requestInterval := make(chan bool, 1)
	go agent.LazyIntervalRange(requestInterval, 2*time.Second, 3*time.Second)

	resultChannel := make(chan agent.Result, limit)

	for {
		select {
		case result := <-resultChannel:
			ok, err := agent.SubmitResult(c, agentId, result.ExpressionID, result.Result)
			if err != nil || !ok {
				return
			}
			delete(workers, result.ExpressionID)

		case <-aliveInterval:
			ok, err := agent.SendAliveMessage(c, agentId)
			if err != nil || !ok {
				return
			}
		case <-requestInterval:
			if len(workers) >= limit {
				continue
			}

			res, err := agent.RequestExpression(c, agentId)
			if err != nil || res.IsDeleted {
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

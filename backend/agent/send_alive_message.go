package agent

import (
	pb "backend/proto"
	"context"
	"fmt"
	"time"
)

func SendAliveMessage(c pb.OrchestratorClient, agentID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var data = &pb.IAmAliveRequest{AgentID: agentID}

	var err error
	var res *pb.IAmAliveResponse

	for range 3 {
		res, err = c.IAmAlive(ctx, data)

		if err != nil {
			fmt.Printf("%v", err)
			time.Sleep(time.Second)
			continue
		}

		return !res.IsDeleted, nil
	}

	return false, err
}

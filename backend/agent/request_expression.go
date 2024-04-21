package agent

import (
	pb "backend/proto"
	"context"
	"fmt"
	"time"
)

func RequestExpression(c pb.OrchestratorClient, agentID string) (*pb.GetExpressionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var data = &pb.GetExpressionRequest{AgentID: agentID}

	var err error
	var res *pb.GetExpressionResponse

	for range 3 {
		res, err = c.GetExpression(ctx, data)

		if err != nil {
			fmt.Printf("%v", err)
			time.Sleep(time.Second)
			continue
		}

		return res, nil
	}

	return nil, err
}

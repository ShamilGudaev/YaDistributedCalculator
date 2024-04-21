package agent

import (
	pb "backend/proto"
	"context"
	"fmt"
	"time"
)

func SubmitResult(c pb.OrchestratorClient, agentID string, expressionID uint64, result float64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var data = &pb.SubmitResultRequest{
		ExpressionID: expressionID,
		AgentID:      agentID,
		Result:       fmt.Sprintf("%g", result),
	}

	var err error
	var res *pb.SubmitResultResponse

	for range 3 {
		res, err = c.SubmitResult(ctx, data)

		if err != nil {
			fmt.Printf("%v", err)
			time.Sleep(time.Second)
			continue
		}

		return res.Accepted, nil
	}

	return false, err
}

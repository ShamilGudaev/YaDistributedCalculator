package agent

import (
	"backend/orchestrator/endpoints/agent"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SubmitResult(agentID string, expressionID uint64, result float64) bool {
	reqData, err := json.Marshal(&agent.SubmitResultRequest{
		ExpressionID: expressionID,
		AgentID:      agentID,
		Result:       result,
	})

	if err != nil {
		panic(err)
	}

	for {
		cont, result := submitResult2(agentID, reqData)
		if cont {
			continue
		}

		return result
	}
}

func submitResult2(agentID string, reqData []byte) (cont bool, result bool) {
	const path = "submit_result"

	res, err := http.Post(
		fmt.Sprintf("http://orchestrator:1324/%s", path),
		"application/json",
		bytes.NewBuffer(reqData),
	)

	if printIfHttpReqFailed(err, &agentID, path, 5*time.Second) {
		return true, false
	}
	defer res.Body.Close()

	panicIfBadRequest(res, &agentID, path)
	if printIfInternalServerError(res, &agentID, path, 5*time.Second) {
		return true, false
	}
	panicIfNotOk(res, &agentID, path)

	var resData agent.SubmitResultResponse
	if printIfResponseIsInvalid(
		json.NewDecoder(res.Body).Decode(&resData),
		&agentID, path, 5*time.Second,
	) {
		return true, false
	}

	return false, resData.Accepted
}

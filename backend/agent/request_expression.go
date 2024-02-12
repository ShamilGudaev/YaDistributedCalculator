package agent

import (
	"backend/orchestrator/endpoints/agent"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func RequestExpression(agentID string) *agent.GetExpressionResponse {
	reqData, err := json.Marshal(&agent.GetExpressionRequest{AgentID: agentID})

	if err != nil {
		panic(err)
	}

	for {
		cont, result := requestExpression2(agentID, reqData)
		if cont {
			continue
		}

		return result
	}
}

func requestExpression2(agentID string, reqData []byte) (cont bool, result *agent.GetExpressionResponse) {
	const path = "get_expression"

	res, err := http.Post(
		fmt.Sprintf("http://orchestrator:1324/%s", path),
		"application/json",
		bytes.NewBuffer(reqData),
	)

	if printIfHttpReqFailed(err, &agentID, path, 5*time.Second) {
		return true, nil
	}
	defer res.Body.Close()

	panicIfBadRequest(res, &agentID, path)
	if printIfInternalServerError(res, &agentID, path, 5*time.Second) {
		return true, nil
	}
	panicIfNotOk(res, &agentID, path)

	var resData agent.GetExpressionResponse
	json.NewDecoder(res.Body).Decode(&resData)

	return false, &resData
}

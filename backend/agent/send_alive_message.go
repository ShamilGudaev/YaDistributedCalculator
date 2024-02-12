package agent

import (
	"backend/orchestrator/endpoints/agent"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SendAliveMessage(agentID string) bool {
	reqData, err := json.Marshal(&agent.IAmAliveRequest{AgentID: agentID})

	if err != nil {
		panic(err)
	}

	for {
		cont, result := sendAliveMessage2(agentID, reqData)
		if cont {
			continue
		}

		return result
	}
}

func sendAliveMessage2(agentID string, reqData []byte) (cont bool, result bool) {
	const path = "i_am_alive"

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

	var resData agent.IAmAliveResponse
	json.NewDecoder(res.Body).Decode(&resData)

	return false, !resData.IsDeleted
}

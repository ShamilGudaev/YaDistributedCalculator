package client

import (
	"backend/orchestrator/db"
	"backend/orchestrator/events"
	"backend/orchestrator/middleware"
	"backend/orchestrator/util"
	"backend/parser"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ExpressionData struct {
	ID      uint64  `json:"id,string"`
	Text    string  `json:"text"`
	Result  *string `json:"result"`
	AgentID *string `json:"agentId"`
}

type AgentsData struct {
	ID            string   `json:"id"`
	ExpressionIDs []string `json:"expressionIds"`
	LastSeen      string   `json:"lastSeen"`
	DeletedAt     *string  `json:"deletedAt"`
}

type ExecutionTimeData struct {
	OpMulMS uint32 `json:"opMulMS"`
	OpDivMS uint32 `json:"opDivMS"`
	OpAddMS uint32 `json:"opAddMS"`
	OpSubMS uint32 `json:"opSubMS"`
}

type InitialData struct {
	Expressions   []ExpressionData  `json:"expressions"`
	Agents        []AgentsData      `json:"agents"`
	ExecutionTime ExecutionTimeData `json:"executionTime"`
}

func Subscribe(c echo.Context) error {
	e := events.EventsEmitter.On("client")
	defer events.EventsEmitter.Off("client", e)
	userID, ok := c.Get(middleware.UserIDKey).(uint64)
	if !ok {
		fmt.Fprintf(c.Response(), "Error parse userID")
	}

	clientEventPath := fmt.Sprintf("client/%d", userID)
	e2 := events.EventsEmitter.On(clientEventPath)
	defer events.EventsEmitter.Off(clientEventPath, e2)

	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "http://127.0.0.1:5173")
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().WriteHeader(http.StatusOK)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var initialData InitialData = InitialData{
			Expressions:   []ExpressionData{},
			Agents:        []AgentsData{},
			ExecutionTime: ExecutionTimeData{},
		}

		var expressions []db.Expression
		res := tx.Find(&expressions, "user_id = ?", userID)
		if err := res.Error; err != nil {
			return err
		}
		for _, expr := range expressions {
			initialData.Expressions = append(initialData.Expressions, ExpressionData{
				ID:      expr.ID,
				Text:    expr.Text,
				Result:  expr.Result,
				AgentID: expr.AgentID,
			})
		}

		var agents []db.Agent
		res = tx.Model(&db.Agent{}).Preload("Expressions").Find(&agents)
		if err := res.Error; err != nil {
			return err
		}
		for _, agent := range agents {
			var exprIds = make([]string, 0)
			for _, expr := range agent.Expressions {
				exprIds = append(exprIds, fmt.Sprintf("%d", expr.ID))
			}

			v := AgentsData{
				ID:            agent.ID,
				ExpressionIDs: exprIds,
				LastSeen:      agent.LastSeen.Format(util.DateFormat),
			}

			if agent.DeletedAt != nil {
				a := agent.DeletedAt.Format(util.DateFormat)
				v.DeletedAt = &a
			}

			initialData.Agents = append(initialData.Agents, v)
		}

		var executionTime []*db.ExecutionTime
		res = tx.Find(&executionTime)
		if err := res.Error; err != nil {
			return err
		}

		for _, execTime := range executionTime {
			switch execTime.Operator {
			case parser.OpMul:
				initialData.ExecutionTime.OpMulMS = execTime.DurationMS
			case parser.OpDiv:
				initialData.ExecutionTime.OpDivMS = execTime.DurationMS
			case parser.OpAdd:
				initialData.ExecutionTime.OpAddMS = execTime.DurationMS
			case parser.OpSub:
				initialData.ExecutionTime.OpSubMS = execTime.DurationMS
			}
		}

		data, err := json.Marshal(initialData)

		if err != nil {
			return err
		}

		fmt.Fprintf(c.Response(), "event: %s\ndata: %s\n\n", "initial_data", string(data))
		c.Response().Flush()

		return nil
	})

	if err != nil {
		return err
	}

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case v := <-e:
			fmt.Fprint(c.Response(), v.String(0))
			c.Response().Flush()
		case v := <-e2:
			fmt.Fprint(c.Response(), v.String(0))
			c.Response().Flush()
		}
	}
}

package agent

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubmitResultRequest struct {
	AgentID      string  `json:"agentId"`
	ExpressionID uint64  `json:"expressionId,string"`
	Result       float64 `json:"result"`
}

type SubmitResultResponse struct {
	Accepted bool `json:"accepted"`
}

func SubmitResult(c echo.Context) error {
	var req SubmitResultRequest
	if err := c.Bind(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var expression db.Expression
		// Пытаемся найти выражение с нужными id
		res := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", req.ExpressionID).
			Where("agent_id = ?", req.AgentID).
			Limit(1).
			Find(&expression)

		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			// Если не нашли, не принимаем
			return c.JSON(http.StatusOK, &SubmitResultResponse{Accepted: false})
		}

		// Если нашли, обновляем
		expression.AgentID = nil
		expression.Result = &req.Result
		if err := tx.Save(&expression).Error; err != nil {
			return err
		}

		c.JSON(http.StatusOK, &SubmitResultResponse{Accepted: true})

		events.SendEventToClients(
			"expressions_change",
			[]client.ExpressionData{
				{
					ID:      expression.ID,
					Text:    expression.Text,
					AgentID: expression.AgentID,
					Result:  expression.Result,
				},
			},
		)

		return nil
	})

	if err != nil {
		c.Logger().Error(err)
		c.String(http.StatusInternalServerError, "")
	}

	return nil
}

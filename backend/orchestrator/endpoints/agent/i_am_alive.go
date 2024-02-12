package agent

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"backend/orchestrator/util"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAmAliveRequest struct {
	AgentID string `json:"agentId"`
}

type IAmAliveResponse struct {
	IsDeleted bool `json:"isDeleted"`
}

func IAmAlive(c echo.Context) error {
	var req IAmAliveRequest
	if err := c.Bind(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid request body")
		return nil
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Пытаемся получить агента
		var agent db.Agent
		res := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", req.AgentID).
			Find(&agent)

		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			// Если не получили, создаем
			agent.ID = req.AgentID
			agent.LastSeen = time.Now()
			res := tx.Save(&agent)
			if err := res.Error; err != nil {
				return err
			}
			return c.JSON(http.StatusOK, &IAmAliveResponse{IsDeleted: false})
		}

		if agent.DeletedAt != nil {
			// Если удален, уведомляем об этом
			return c.JSON(http.StatusOK, &IAmAliveResponse{IsDeleted: true})
		}

		// Обновляем LastSeen
		agent.LastSeen = time.Now()
		res = tx.Save(&agent)
		if err := res.Error; err != nil {
			return err
		}

		err := c.JSON(http.StatusOK, &IAmAliveResponse{IsDeleted: false})

		if err != nil {
			return nil
		}

		var expressionIds []uint64

		res = tx.
			Model(&db.Expression{}).
			Where("agent_id = ?", agent.ID).
			Pluck("id", &expressionIds)

		if err := res.Error; err != nil {
			return err
		}

		var expressionIdsStr = make([]string, len(expressionIds))
		for i, id := range expressionIds {
			expressionIdsStr[i] = fmt.Sprintf("%d", id)
		}

		events.SendEventToClients("agents_change", []client.AgentsData{{
			ID:            agent.ID,
			ExpressionIDs: expressionIdsStr,
			LastSeen:      agent.LastSeen.Format(util.DateFormat)},
		})

		return nil
	})

	if err != nil {
		c.Logger().Error(err)
		c.String(http.StatusInternalServerError, "")
	}

	return nil
}

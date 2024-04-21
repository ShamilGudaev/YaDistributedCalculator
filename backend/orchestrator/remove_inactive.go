package orchestrator

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"backend/orchestrator/util"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RemoveInactive() {
	time.Sleep(30 * time.Second)
	for {
		now := time.Now()
		deleteOld(now)
		markAsDeleted(now)
		time.Sleep(5 * time.Second)
	}
}

func deleteOld(now time.Time) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		return deleteOld0(tx, now)
	})
	if err != nil {
		println(err.Error())
	}
}

func deleteOld0(tx *gorm.DB, now time.Time) error {
	// Получаем те, которые были удалены больше 10 минут назад
	var completelyDeleted []db.Agent
	err := tx.
		Clauses(clause.Returning{}).
		Where("deleted_at < ?", now.Add(-10*time.Minute)).
		Delete(&completelyDeleted).
		Error

	if err != nil {
		return err
	}

	if len(completelyDeleted) == 0 {
		return nil
	}

	completelyDeletedIds := make([]string, len(completelyDeleted))
	for i, agent := range completelyDeleted {
		completelyDeletedIds[i] = agent.ID
	}

	events.SendEventToClients("agents_remove", &completelyDeletedIds)
	return nil
}

func markAsDeleted(now time.Time) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		return markAsDeleted0(tx, now)
	})
	if err != nil {
		println(err.Error())
	}
}

func markAsDeleted0(tx *gorm.DB, now time.Time) error {
	var markedAgents []db.Agent
	err := tx.
		Model(&markedAgents).
		Clauses(clause.Returning{}).
		Where("last_seen < ?", now.Add(-time.Minute)).
		Where("deleted_at IS NULL").
		Update("deleted_at", now).
		Error

	if err != nil {
		return err
	}

	if len(markedAgents) == 0 {
		return nil
	}

	markedAgentIds := make([]string, len(markedAgents))
	for i, a := range markedAgents {
		markedAgentIds[i] = a.ID
	}

	var expressions []db.Expression

	err = tx.
		Model(&expressions).
		Clauses(clause.Returning{}).
		Where("agent_id IN ?", markedAgentIds).
		Update("agent_id", nil).
		Error

	if err != nil {
		return err
	}

	agentsData := make([]client.AgentsData, len(markedAgents))
	for i, agent := range markedAgents {
		deletedAt := agent.DeletedAt.Format(util.DateFormat)
		agentsData[i] = client.AgentsData{
			ID:            agent.ID,
			ExpressionIDs: []string{},
			LastSeen:      agent.LastSeen.Format(util.DateFormat),
			DeletedAt:     &deletedAt,
		}
	}

	events.SendEventToClients("agents_change", agentsData)

	if len(expressions) == 0 {
		return nil
	}

	expressionsData := make(map[uint64][]client.ExpressionData, len(expressions))

	for _, expression := range expressions {
		var userExpressions = expressionsData[expression.UserID]
		if userExpressions == nil {
			userExpressions = make([]client.ExpressionData, 0, 1)
		}
		userExpressions = append(userExpressions, client.ExpressionData{
			ID:      expression.ID,
			Text:    expression.Text,
			Result:  expression.Result,
			AgentID: expression.AgentID,
		})
		expressionsData[expression.UserID] = userExpressions
	}

	for userID, data := range expressionsData {
		events.SendEventToClientByUserID(userID, "expressions_change", data)
	}

	return nil
}

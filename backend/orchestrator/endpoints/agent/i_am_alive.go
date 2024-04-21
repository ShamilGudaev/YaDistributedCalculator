package agent

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"backend/orchestrator/util"
	"context"
	"fmt"
	"time"

	pb "backend/proto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *AgentGRPCServer) IAmAlive(ctx context.Context, req *pb.IAmAliveRequest) (outRes *pb.IAmAliveResponse, outErr error) {
	outErr = db.DB.Transaction(func(tx *gorm.DB) error {
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

			outRes = &pb.IAmAliveResponse{IsDeleted: false}
			return nil
		}

		if agent.DeletedAt != nil {
			// Если удален, уведомляем об этом
			outRes = &pb.IAmAliveResponse{IsDeleted: true}
			return nil
		}

		// Обновляем LastSeen
		agent.LastSeen = time.Now()
		res = tx.Save(&agent)
		if err := res.Error; err != nil {
			return err
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

		outRes = &pb.IAmAliveResponse{IsDeleted: false}
		return nil
	})

	return
}

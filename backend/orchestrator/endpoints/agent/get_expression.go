package agent

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"backend/orchestrator/util"
	"backend/parser"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	pb "backend/proto"
)

// type GetExpressionRequest struct {
// 	AgentID string `json:"agentId"`
// }

// type GetExpressionResponse struct {
// 	Data      *GetExpressionResponseData `json:"data"`
// 	IsDeleted bool                       `json:"isDeleted"`
// }

// type GetExpressionResponseData struct {
// 	ExpressionID uint64 `json:"expressionId,string"`
// 	Expression   string `json:"expression"`
// 	OpMulMS      uint32 `json:"opMulMS"`
// 	OpDivMS      uint32 `json:"opDivMS"`
// 	OpAddMS      uint32 `json:"opAddMS"`
// 	OpSubMS      uint32 `json:"opSubMS"`
// }

func (s *AgentGRPCServer) GetExpression(ctx context.Context, req *pb.GetExpressionRequest) (outRes *pb.GetExpressionResponse, outErr error) {
	outErr = db.DB.Transaction(func(tx *gorm.DB) error {
		var agent db.Agent

		// Пытаемся найти агента в бд
		res := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", req.AgentID).
			Limit(1).
			Find(&agent)

		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			// Если не нашли, создаем
			agent = db.Agent{
				ID:       req.AgentID,
				LastSeen: time.Now(),
			}

			res = tx.
				Clauses(clause.Returning{}).
				Create(&agent)

			if err := res.Error; err != nil {
				return err
			}

		} else if agent.DeletedAt != nil {
			// Если агент удален, уведомляем об этом (агент должен сам сменить id)
			outRes = &pb.GetExpressionResponse{IsDeleted: true}
			return nil
		}

		// Пытаемся найти выражение без агента
		var expression db.Expression
		res = tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("result IS NULL").
			Where("agent_id IS NULL").
			Limit(1).
			Find(&expression)

		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			// Если не нашли, уведомляем об этом
			outRes = &pb.GetExpressionResponse{
				IsDeleted: false,
				Data:      nil,
			}
			return nil
		}

		// Нашли, задаем агента для выражения
		expression.AgentID = &req.AgentID
		res = tx.Save(&expression)

		if err := res.Error; err != nil {
			return err
		}

		var data = &pb.GetExpressionResponse_GetExpressionResponseData{
			ExpressionID: expression.ID,
			Expression:   expression.Text,
		}

		// Получаем настройки продолжительности
		var executionTime []*db.ExecutionTime
		res = tx.Find(&executionTime)
		if err := res.Error; err != nil {
			return err
		}

		for _, execTime := range executionTime {
			switch execTime.Operator {
			case parser.OpMul:
				data.OpMulMS = execTime.DurationMS
			case parser.OpDiv:
				data.OpDivMS = execTime.DurationMS
			case parser.OpAdd:
				data.OpAddMS = execTime.DurationMS
			case parser.OpSub:
				data.OpSubMS = execTime.DurationMS
			}
		}

		// Получаем полную информацию об агенте
		res = tx.
			Model(&db.Agent{}).
			Preload("Expressions").
			First(&agent)

		if err := res.Error; err != nil {
			return err
		}

		// Отправляем выражение
		outRes = &pb.GetExpressionResponse{
			Data:      data,
			IsDeleted: false,
		}

		events.SendEventToClientByUserID(
			expression.UserID,
			"expressions_change",
			[]client.ExpressionData{
				{
					ID:      expression.ID,
					Text:    expression.Text,
					AgentID: expression.AgentID,
				},
			},
		)

		expressionIds := make([]string, len(agent.Expressions))
		for i, expr := range agent.Expressions {
			expressionIds[i] = fmt.Sprintf("%d", expr.ID)
		}

		events.SendEventToClients(
			"agents_change",
			[]client.AgentsData{
				{
					ID:            agent.ID,
					ExpressionIDs: expressionIds,
					LastSeen:      agent.LastSeen.Format(util.DateFormat),
					DeletedAt:     nil,
				},
			},
		)
		return nil
	})

	return
}

package agent

import (
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/events"
	"context"

	pb "backend/proto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *AgentGRPCServer) SubmitResult(ctx context.Context, req *pb.SubmitResultRequest) (outRes *pb.SubmitResultResponse, outErr error) {

	outErr = db.DB.Transaction(func(tx *gorm.DB) error {
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
			outRes = &pb.SubmitResultResponse{Accepted: false}
			return nil
		}

		// Если нашли, обновляем
		expression.AgentID = nil
		expression.Result = &req.Result
		if err := tx.Save(&expression).Error; err != nil {
			return err
		}

		outRes = &pb.SubmitResultResponse{Accepted: true}

		events.SendEventToClientByUserID(expression.UserID, "expressions_change", []client.ExpressionData{
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

	return
}

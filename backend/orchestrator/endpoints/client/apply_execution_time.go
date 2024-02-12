package client

import (
	"backend/orchestrator/db"
	"backend/orchestrator/events"
	"backend/parser"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ApplyExecutionTime(c echo.Context) error {
	req := new(ExecutionTimeData)

	if err := c.Bind(req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&db.ExecutionTime{}).Where("operator = ?", parser.OpMul).UpdateColumn("duration_ms", req.OpMulMS).Error; err != nil {
			return err
		}
		if err := tx.Model(&db.ExecutionTime{}).Where("operator = ?", parser.OpDiv).UpdateColumn("duration_ms", req.OpDivMS).Error; err != nil {
			return err
		}
		if err := tx.Model(&db.ExecutionTime{}).Where("operator = ?", parser.OpAdd).UpdateColumn("duration_ms", req.OpAddMS).Error; err != nil {
			return err
		}
		if err := tx.Model(&db.ExecutionTime{}).Where("operator = ?", parser.OpSub).UpdateColumn("duration_ms", req.OpSubMS).Error; err != nil {
			return err
		}
		events.SendEventToClients("exec_time_change", req)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.String(http.StatusOK, "{\"ok\":true}")
		return nil
	})

	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return nil
	}

	return nil
}

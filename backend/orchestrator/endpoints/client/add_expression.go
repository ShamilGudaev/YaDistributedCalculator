package client

import (
	"backend/orchestrator/db"
	"backend/orchestrator/events"
	"backend/parser"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AddExpressionRequest struct {
	Expression string `json:"expression"`
}

type AddExpressionResponse struct {
	ExpressionID *uint64 `json:"expressionId,string"`
	Error        *string `json:"error"`
}

func AddExpression(c echo.Context) error {
	req := new(AddExpressionRequest)
	if err := c.Bind(req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	// Проверяем валидность
	_, err := parser.Parser.ParseString("", req.Expression)
	if err != nil {
		error := err.Error()
		c.JSON(http.StatusOK, &AddExpressionResponse{
			Error: &error,
		})
		return nil
	}

	expression := db.Expression{
		Text: req.Expression,
	}

	result := db.DB.Create(&expression)

	if result.Error != nil {
		c.Logger().Error(result.Error)
		c.String(http.StatusInternalServerError, "Internal server error")
		return nil
	}

	c.JSON(http.StatusOK, &AddExpressionResponse{
		ExpressionID: &expression.ID,
	})

	events.SendEventToClients("expressions_change", [1]ExpressionData{{ID: expression.ID, Text: expression.Text}})

	return nil
}

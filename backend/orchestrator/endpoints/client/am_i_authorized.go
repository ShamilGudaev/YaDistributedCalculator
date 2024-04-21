package client

import (
	"backend/orchestrator/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AmIAuthorized(c echo.Context) error {
	_, ok := c.Get(middleware.UserIDKey).(uint64)
	c.JSON(http.StatusOK, ok)
	return nil
}

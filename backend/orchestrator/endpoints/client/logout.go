package client

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{Name: "token", Value: "", HttpOnly: true, MaxAge: -1})
	c.String(http.StatusOK, "")
	return nil
}

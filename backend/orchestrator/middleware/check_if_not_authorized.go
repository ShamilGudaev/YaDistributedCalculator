package middleware

import "github.com/labstack/echo/v4"

func CheckIfNotAuthorized(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := c.Get(UserIDKey).(uint64)
		if !ok {
			return next(c)
		}

		return c.String(echo.ErrForbidden.Code, "Already authorized")
	}
}

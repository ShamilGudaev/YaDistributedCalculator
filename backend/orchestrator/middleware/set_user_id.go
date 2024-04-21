package middleware

import (
	"backend/orchestrator/cfg"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTUser struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

var validValidationAlgs = []string{"HS256"}

const UserIDKey = "user_id"

// Middleware
func SetUserIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("token")
		if err != nil {
			return next(c)
		}

		tokenString := cookie.Value

		token, err := jwt.ParseWithClaims(tokenString, &JWTUser{},
			func(token *jwt.Token) (interface{}, error) {
				return cfg.JwtSecret, nil
			}, jwt.WithValidMethods(validValidationAlgs))

		if err != nil {
			return next(c)
		}

		if !token.Valid {
			return c.String(http.StatusUnauthorized, "Invalid token")
		}

		claims, ok := token.Claims.(*JWTUser)
		if !ok {
			return c.String(http.StatusUnauthorized, "Parse token")
		}

		userID, err := strconv.ParseUint(claims.ID, 10, 64)

		if err != nil {
			return c.String(http.StatusUnauthorized, "Invalid token")
		}

		c.Set(UserIDKey, userID)
		return next(c)
	}
}

package client

import (
	"backend/orchestrator/cfg"
	"backend/orchestrator/db"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Result string `json:"result"` // "authorized" | "invalid_credentials"
}

type JWTUser struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

var regex = regexp.MustCompile(`^[a-zA-Z0-9_-]{4,20}$`)

func UserAuthorization(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	if !regex.MatchString(req.Login) {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	req.Login = strings.ToLower(req.Login)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var user db.User
		res := tx.
			Where("login = ?", req.Login).
			Find(&user)

		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			return c.JSON(http.StatusOK, &AuthResponse{Result: "invalid_credentials"})
		}

		// passBytes := []byte(fmt.Sprintf("%s%s", req.Password, user.Salt))

		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusOK, &AuthResponse{Result: "invalid_credentials"})
			return nil
		}

		jwtData := JWTUser{
			fmt.Sprintf("%d", user.ID),
			jwt.RegisteredClaims{},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtData)

		tokenString, err := token.SignedString(cfg.JwtSecret)
		if err != nil {
			return err
		}

		cookie := new(http.Cookie)
		cookie.Name = "token"
		cookie.Value = tokenString
		cookie.Expires = time.Now().Add(24 * 365 * time.Hour)
		cookie.HttpOnly = true
		c.SetCookie(cookie)
		c.JSON(http.StatusOK, &AuthResponse{Result: "authorized"})

		return nil
	})

	if err != nil {
		c.Logger().Error(err.Error())
		c.String(http.StatusInternalServerError, "")
	}
	return nil
}

package client

import (
	"backend/orchestrator/cfg"
	"backend/orchestrator/db"
	"fmt"

	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserResponse struct {
	Result string `json:"result"` // "registered" | "already_exists"
}

func UserRegistration(c echo.Context) error {
	var req UserRequest
	if err := c.Bind(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	if !regex.MatchString(req.Login) {
		c.String(http.StatusBadRequest, "Invalid request body")
		return nil
	}

	//Конструируем хеш пароля
	req.Login = strings.ToLower(req.Login)

	/*saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		fmt.Println("Error generating salt:", err)
		return err
	}
	saltString := base64.StdEncoding.EncodeToString(saltBytes)
	passBytes := []byte(fmt.Sprintf("%s%s", req.Password, saltString)) // перегнать в байты*/

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // байты посоленого пароля сюда, а не байты соли, лол
	if err != nil {
		return err
	}

	passwordHash := string(hashedBytes)

	//Пытаемся добавить пользователя в базу
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		user := db.User{
			Login:        req.Login,
			PasswordHash: passwordHash,
		}

		res := tx.
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&user)

		if err := res.Error; err != nil {
			return err
		}

		// Если не удалось добавить пользователя
		if res.RowsAffected == 0 {
			c.JSON(http.StatusOK, &UserResponse{Result: "already_exists"})
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
		c.JSON(http.StatusOK, &UserResponse{Result: "registered"})

		return nil
	})

	if err != nil {
		c.Logger().Error(err.Error())
		c.String(http.StatusInternalServerError, "")
	}

	return nil
}

// json запрос
// login string
// password string
// проверяет, есть ли пользователь
// если есть, пишет ошибку protobuf { error: string }
// https://pkg.go.dev/golang.org/x/crypto/argon2
// если нет, то добавить в бд в хэшем Argon2id 19 MiB, 2 итерации, 1 степень параллельности
// и отправляем jwt http-only cookie с именем token, куда пишем id пользователя
// для jwt (HS512) нужно сгенерировать длинную строку (salt), передавать ее через docker secrets
// добавить secrets в .gitignore
// из /run/secrets/jwt_salt

// json запрос
// login string
// password string
// проверяет, есть ли пользователь
// если есть, то отправляем jwt http-only cookie с именем token, куда пишем id пользователя
// если нет, то пишем ошибку

//golang.org/x/crypto

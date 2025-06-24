// Package middleware предоставляет Gin-middleware для аутентификации с помощью JWT.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Claims расширяет jwt.RegisteredClaims, добавляя UserID –
// уникальный идентификатор пользователя.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// cookieName — имя cookie, в котором хранится JWT.
const cookieName = "userID"

// AuthMiddleware возвращает Gin-мiddleware, который:
//  1. проверяет наличие и валидность JWT в cookie с именем cookieName;
//  2. при отсутствии или невалидном токене создаёт новый, устанавливает его в cookie
//     и сохраняет userID в контексте запроса;
//  3. при валидном токене извлекает userID из него и сохраняет в контексте.
func AuthMiddleware(secretKey string, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			logger.Debug("JWT cookie not found, generating new one")
			issueNewToken(c, secretKey, logger)
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			logger.Debugw("Invalid JWT, issuing new one", "error", err)
			issueNewToken(c, secretKey, logger)
			return
		}

		c.Set(cookieName, claims.UserID)
	}
}

// issueNewToken генерирует новый JWT для уникального userID,
// устанавливает его в cookie и сохраняет userID в контексте.
func issueNewToken(c *gin.Context, secretKey string, logger *zap.SugaredLogger) {
	userID := uuid.New().String()

	tokenString, err := generateJWT(userID, secretKey)
	if err != nil {
		logger.Errorw("Failed to generate JWT", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		HttpOnly: true,
	})

	c.Set(cookieName, userID)
}

// generateJWT создаёт и подписывает JWT для заданного userID и секрета.
func generateJWT(userID, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})

	return token.SignedString([]byte(secret))
}

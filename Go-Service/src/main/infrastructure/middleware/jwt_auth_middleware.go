package middleware

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware handles JWT authentication and authorization
func JWTAuthMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Cookie first (preferred method)
		cookie, err := c.Cookie("token")
		if err == nil && cookie != "" {
			tokenString = cookie
		} else {
			// Fallback to Authorization header for backward compatibility
			tokenString = c.GetHeader("Authorization")
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing token"})
				c.Abort()
				return
			}
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		claims := &dto.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		// Store claims in context for later use
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

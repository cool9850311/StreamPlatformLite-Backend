package middleware

import (
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"context"
	"net/http"

	claims "github.com/cool9850311/StreamPlatformLite-Core/pkg/claims"
	"github.com/cool9850311/StreamPlatformLite-Core/pkg/csrf"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware handles JWT authentication and authorization
func JWTAuthMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil || tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing token"})
			c.Abort()
			return
		}

		cl := &claims.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, cl, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		// Store claims in context for later use
		ctx := context.WithValue(c.Request.Context(), "claims", cl)
		c.Request = c.Request.WithContext(ctx)

		// Also store user_id in Gin context for rate limiting middleware
		c.Set("user_id", cl.UserID)

		// CSRF: enforce on all state-changing methods
		safeMethod := c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS"
		if !safeMethod {
			csrfToken := c.GetHeader("X-XSRF-TOKEN")
			if csrfToken == "" || !csrf.ValidateCsrfToken(csrfToken, config.AppConfig.JWT.SecretKey, cl.UserID) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid CSRF token"})
				return
			}
		}

		c.Next()
	}
}

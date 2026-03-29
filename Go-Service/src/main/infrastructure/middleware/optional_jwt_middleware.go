package middleware

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OptionalJWTAuthMiddleware 可选的JWT认证中间件
// 如果有有效token，解析并存入context
// 如果没有token或token无效，设置为Anonymous角色
func OptionalJWTAuthMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. 尝试从Cookie读取
		cookie, err := c.Cookie("token")
		if err == nil && cookie != "" {
			tokenString = cookie
		} else {
			// 2. 尝试从Authorization Header读取
			tokenString = c.GetHeader("Authorization")
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		// 如果没有token，设置为Anonymous并继续
		if tokenString == "" {
			claims := &dto.Claims{
				Role: role.Anonymous,
			}
			ctx := context.WithValue(c.Request.Context(), "claims", claims)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		// 3. 尝试解析JWT
		claims := &dto.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.SecretKey), nil
		})

		// 如果token无效，设置为Anonymous
		if err != nil || !token.Valid {
			logger.Warn(c.Request.Context(), "Invalid JWT token, treating as anonymous: "+err.Error())
			claims = &dto.Claims{
				Role: role.Anonymous,
			}
			ctx := context.WithValue(c.Request.Context(), "claims", claims)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		// Token有效，存入context
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

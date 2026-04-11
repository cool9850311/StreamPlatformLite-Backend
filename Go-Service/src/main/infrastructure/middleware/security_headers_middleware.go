package middleware

import (
	"Go-Service/src/main/infrastructure/config"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	// HSTS: 仅在生产环境 HTTPS 时启用 (STSSeconds > 0 才会添加 HSTS 头)
	stsSeconds := int64(0)
	if config.AppConfig.Server.HTTPS {
		stsSeconds = 31536000 // 1 year (31536000 seconds)
	}

	secureConfig := secure.Config{
		// X-Frame-Options: DENY (防止 Clickjacking)
		FrameDeny: true,

		// X-Content-Type-Options: nosniff (防止 MIME 嗅探)
		ContentTypeNosniff: true,

		// X-XSS-Protection: 1; mode=block
		BrowserXssFilter: true,

		// Referrer-Policy: strict-origin-when-cross-origin
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// HSTS: 只在 HTTPS 启用时设置 (STSSeconds > 0)
		STSSeconds:           stsSeconds,
		STSIncludeSubdomains: true,
		STSPreload:           false,

		// 不使用 IsDevelopment，以确保所有安全头都生效
		IsDevelopment: false,

		// Zero-trust API CSP: prevents API responses from being treated as HTML
		// and prevents embedding in iframes
		ContentSecurityPolicy: "default-src 'none'; frame-ancestors 'none'",
	}

	secureMiddleware := secure.New(secureConfig)

	// Return a wrapper that adds Permissions-Policy header manually
	return func(c *gin.Context) {
		// Add Permissions-Policy header (newer standard replacing Feature-Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Call the secure middleware
		secureMiddleware(c)
	}
}

package infrastructure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type mockLogger struct{}

func (m *mockLogger) Panic(ctx context.Context, msg string) {}
func (m *mockLogger) Fatal(ctx context.Context, msg string) {}
func (m *mockLogger) Error(ctx context.Context, msg string) {}
func (m *mockLogger) Warn(ctx context.Context, msg string)  {}
func (m *mockLogger) Info(ctx context.Context, msg string)  {}
func (m *mockLogger) Debug(ctx context.Context, msg string) {}
func (m *mockLogger) Trace(ctx context.Context, msg string) {}

const testSecret = "m6zr8Z1NL3ctUi2lcF8QEtZxI"

func init() {
	config.AppConfig.JWT.SecretKey = testSecret
	gin.SetMode(gin.TestMode)
}

func generateTestJWT(userID string) string {
	claims := &dto.Claims{
		UserID:   userID,
		UserName: "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	return token
}

func setupTestRouter(log logger.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.JWTAuthMiddleware(log))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	return r
}

func TestJWTMiddleware_CSRF_MissingHeader(t *testing.T) {
	log := &mockLogger{}
	r := setupTestRouter(log)

	jwtToken := generateTestJWT("test-user")
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["message"] != "invalid CSRF token" {
		t.Fatalf("expected 'invalid CSRF token', got %q", body["message"])
	}
}

func TestJWTMiddleware_CSRF_InvalidToken(t *testing.T) {
	log := &mockLogger{}
	r := setupTestRouter(log)

	jwtToken := generateTestJWT("test-user")
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	req.Header.Set("X-XSRF-TOKEN", "bad")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestJWTMiddleware_CSRF_WrongUser(t *testing.T) {
	log := &mockLogger{}
	r := setupTestRouter(log)

	jwtToken := generateTestJWT("test-user")
	csrfToken, _ := util.GenerateCsrfToken(testSecret, "other-user")
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	req.Header.Set("X-XSRF-TOKEN", csrfToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestJWTMiddleware_CSRF_ValidToken(t *testing.T) {
	log := &mockLogger{}
	r := setupTestRouter(log)

	jwtToken := generateTestJWT("test-user")
	csrfToken, _ := util.GenerateCsrfToken(testSecret, "test-user")
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	req.Header.Set("X-XSRF-TOKEN", csrfToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestJWTMiddleware_CSRF_GetExempt(t *testing.T) {
	log := &mockLogger{}
	r := setupTestRouter(log)

	jwtToken := generateTestJWT("test-user")
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}


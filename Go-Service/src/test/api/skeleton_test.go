// Go-Service/src/test/api/skeleton_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"Go-Service/src/main/domain/entity"
	"Go-Service/src/main/infrastructure/initializer"
	infra_entity "Go-Service/src/main/infrastructure/repository/entity"
	"Go-Service/src/main/infrastructure/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testToken string

func setup() *gin.Engine {
	initializer.InitLog()
	initializer.InitConfig()
	initializer.InitMongoClient()
	initializer.InitLiveStreamService(initializer.Log, initializer.DB)

	r := router.NewRouter(initializer.DB, initializer.Log, initializer.LiveStreamService)

	// Create test user
	testUser := infra_entity.User{
		Username: "testuser",
		Password: "$2a$12$mRejCE/.ZWkISWh6bx9A9eON/hWowfhRaHUzP4/0uH7H4SY602kQG",
		Role:     2,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	insertResult, err := initializer.DB.Collection("users").InsertOne(ctx, testUser)
	if err != nil {
		log.Fatalf("Failed to insert test user: %v", err)
	}

	// Get token
	w := httptest.NewRecorder()
	loginData := `{"username":"testuser","password":"testpass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(loginData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var tokenResponse struct {
		Token string `json:"token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tokenResponse)
	testToken = tokenResponse.Token

	// Remove test user
	_, err = initializer.DB.Collection("users").DeleteOne(ctx, bson.M{"_id": insertResult.InsertedID})
	if err != nil {
		log.Fatalf("Failed to remove test user: %v", err)
	}

	return r
}

func teardown() {
	initializer.CleanupMongo()
}

func TestGetSkeleton(t *testing.T) {
	r := setup()
	defer teardown()

	// Insert test data
	testSkeleton := entity.Skeleton{ID: "1", Name: "Test Skeleton"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := initializer.DB.Collection("skeletons").InsertOne(ctx, testSkeleton)
	assert.Nil(t, err)

	// Test with token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/skeletons/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var skeleton entity.Skeleton
	err = json.Unmarshal(w.Body.Bytes(), &skeleton)
	assert.Nil(t, err)
	assert.Equal(t, "Test Skeleton", skeleton.Name)

	// Test without token (should fail)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/skeletons/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSkeleton(t *testing.T) {
	r := setup()
	defer teardown()

	skeleton := entity.Skeleton{ID: "1", Name: "New Skeleton"}
	jsonSkeleton, _ := json.Marshal(skeleton)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/skeletons", bytes.NewBuffer(jsonSkeleton))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/skeletons/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var createdSkeleton entity.Skeleton
	err := json.Unmarshal(w.Body.Bytes(), &createdSkeleton)
	assert.Nil(t, err)
	assert.Equal(t, "New Skeleton", createdSkeleton.Name)
}

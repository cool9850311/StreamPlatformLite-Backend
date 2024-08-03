// Go-Service/src/test/api/skeleton_api_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Go-Service/src/main/domain/entity"
	"Go-Service/src/main/infrastructure/initializer"
	"Go-Service/src/main/infrastructure/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func setup() *gin.Engine {
	initializer.InitLog()
	initializer.InitConfig()
	initializer.InitMongoClient()

	r := router.NewRouter(initializer.DB, initializer.Log)
	return r
}

func teardown() {
	initializer.CleanupMongo()
}

func TestGetSkeleton(t *testing.T) {
	r := setup()
	defer teardown()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/skeletons/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	initializer.DB.Collection("skeletons").InsertOne(ctx, bson.M{"_id": "1", "name": "Test Skeleton"})

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/skeletons/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var skeleton entity.Skeleton
	err := json.Unmarshal(w.Body.Bytes(), &skeleton)
	assert.Nil(t, err)
	assert.Equal(t, "Test Skeleton", skeleton.Name)
}

func Test_CreateSkeleton(t *testing.T) {
	r := setup()
	defer teardown()

	skeleton := entity.Skeleton{ID: "1", Name: "New Skeleton"}
	jsonSkeleton, _ := json.Marshal(skeleton)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/skeletons", bytes.NewBuffer(jsonSkeleton))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/skeletons/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var createdSkeleton entity.Skeleton
	err := json.Unmarshal(w.Body.Bytes(), &createdSkeleton)
	assert.Nil(t, err)
	assert.Equal(t, "New Skeleton", createdSkeleton.Name)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/interactions", handleInteraction)
	r.GET("/recommendations", handleRecommendations)
	return r
}

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env.test")
	os.Exit(m.Run())
}

func TestInteractionAndRecommendation(t *testing.T) {
	r := setupRouter()

	fmt.Println("TestInteractionAndRecommendation")
	fmt.Println("Embedding API URL: ", os.Getenv("EMBEDDING_API_URL"))
	fmt.Println("Vector DB URL: ", os.Getenv("VECTOR_DB_URL"))
	fmt.Println("PORT: ", os.Getenv("PORT"))
	fmt.Println("Embedding server: ", os.Getenv("EMBEDDING_SERVER_URL"))

	// Test POST /interactions
	body := map[string]string{"user_id": "testuser", "content": "test content"}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/interactions", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test GET /recommendations
	req2, _ := http.NewRequest("GET", "/recommendations?user_id=testuser", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

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
	r.POST("/profile", handleProfile)
	r.GET("/eval", handleEval)
	return r
}

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env.test")
	os.Exit(m.Run())
}

func TestInteractionAndRecommendation(t *testing.T) {
	r := setupRouter()

	// Test POST /interactions
	body := map[string]string{"user_id": "testuser", "content": "test content"}
	b, _ := json.Marshal(body)
	fmt.Printf("[TEST] Enviando POST /interactions: %s\n", string(b))
	req, _ := http.NewRequest("POST", "/interactions", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Printf("[TEST] Status /interactions: %d\n", w.Code)
	fmt.Printf("[TEST] Resposta /interactions: %s\n", w.Body.String())
	assert.Equal(t, 200, w.Code)

	// Test GET /recommendations
	fmt.Println("[TEST] Enviando GET /recommendations?user_id=testuser")
	req2, _ := http.NewRequest("GET", "/recommendations?user_id=testuser", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	fmt.Printf("[TEST] Status /recommendations: %d\n", w2.Code)
	fmt.Printf("[TEST] Resposta /recommendations: %s\n", w2.Body.String())
	assert.Equal(t, 200, w2.Code)

	// Test POST /profile
	profile := map[string]interface{}{"user_id": "testuser", "age": 30, "gender": "F", "location": "SP"}
	bProfile, _ := json.Marshal(profile)
	fmt.Printf("[TEST] Enviando POST /profile: %s\n", string(bProfile))
	reqProfile, _ := http.NewRequest("POST", "/profile", bytes.NewBuffer(bProfile))
	reqProfile.Header.Set("Content-Type", "application/json")
	wProfile := httptest.NewRecorder()
	r.ServeHTTP(wProfile, reqProfile)
	fmt.Printf("[TEST] Status /profile: %d\n", wProfile.Code)
	fmt.Printf("[TEST] Resposta /profile: %s\n", wProfile.Body.String())
	assert.Equal(t, 200, wProfile.Code)

	// // Test GET /eval
	fmt.Println("[TEST] Enviando GET /eval?user_id=testuser&k=5")
	reqEval, _ := http.NewRequest("GET", "/eval?user_id=testuser&k=5", nil)
	wEval := httptest.NewRecorder()
	r.ServeHTTP(wEval, reqEval)
	fmt.Printf("[TEST] Status /eval: %d\n", wEval.Code)
	fmt.Printf("[TEST] Resposta /eval: %s\n", wEval.Body.String())
	assert.Equal(t, 200, wEval.Code)
}

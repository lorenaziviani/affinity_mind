package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Interaction struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type EmbedRequest struct {
	Text string `json:"text"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

type InsertRequest struct {
	ID     string    `json:"id"`
	Vector []float32 `json:"vector"`
}

type QueryRequest struct {
	Vector []float32 `json:"vector"`
	K      int       `json:"k"`
}

type QueryResponse struct {
	IDs       []string  `json:"ids"`
	Distances []float32 `json:"distances"`
}

var (
	embeddingAPI = os.Getenv("EMBEDDING_API_URL")
	vectorDBAPI  = os.Getenv("VECTOR_DB_URL")
)

func main() {
	r := gin.Default()

	r.POST("/interactions", handleInteraction)
	r.GET("/recommendations", handleRecommendations)

	r.Run(":" + os.Getenv("PORT"))
}

func handleInteraction(c *gin.Context) {
	var inter Interaction
	if err := c.ShouldBindJSON(&inter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emb, err := getEmbedding(inter.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding failed"})
		return
	}

	insert := InsertRequest{ID: inter.UserID, Vector: emb}
	if err := postJSON(vectorDBAPI+"/insert", insert, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vector db insert failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "interaction stored"})
}

func handleRecommendations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	// Search for user embedding
	// (Here, for simplicity, it is assumed that the user's embedding is equal to the last sent content)
	// In production, one can maintain history and calculate averages, etc.

	// For example, we will search for the embedding of the user_id itself (ideal: search in the database)
	// Here, for demonstration purposes, we will generate the embedding of the user_id itself
	emb, err := getEmbedding(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding failed"})
		return
	}

	query := QueryRequest{Vector: emb, K: 5}
	var resp QueryResponse
	if err := postJSON(vectorDBAPI+"/query", query, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vector db query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": resp.IDs, "distances": resp.Distances})
}

func getEmbedding(text string) ([]float32, error) {
	req := EmbedRequest{Text: text}
	var resp EmbedResponse
	if err := postJSON(embeddingAPI+"/embed", req, &resp); err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

func postJSON(url string, payload interface{}, out interface{}) error {
	b, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		data, _ := ioutil.ReadAll(resp.Body)
		return json.Unmarshal(data, out)
	}
	return nil
}

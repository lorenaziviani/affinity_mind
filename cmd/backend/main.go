package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	ElapsedMs float64   `json:"elapsed_ms"`
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

type UserProfile struct {
	UserID   string `json:"user_id"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Location string `json:"location"`
}

var userProfiles = make(map[string]UserProfile)

var (
	embeddingLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "embedding_latency_ms",
		Help:    "Embedding API latency in milliseconds",
		Buckets: prometheus.LinearBuckets(10, 10, 10),
	})
	rankingLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ranking_latency_ms",
		Help:    "Ranking (vector DB query) latency in milliseconds",
		Buckets: prometheus.LinearBuckets(10, 10, 10),
	})
)

func init() {
	prometheus.MustRegister(embeddingLatency)
	prometheus.MustRegister(rankingLatency)
}

func getEmbeddingAPI() string {
	return os.Getenv("EMBEDDING_API_URL")
}

func getVectorDBAPI() string {
	return os.Getenv("VECTOR_DB_URL")
}

var userInteractions = make(map[string][]Interaction)

func startEmbeddingRefresher() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			<-ticker.C
			log.Println("[Job] Iniciating embedding refresh...")
			for userID, interactions := range userInteractions {
				if len(interactions) == 0 {
					continue
				}
				last := interactions[len(interactions)-1]
				profile, ok := userProfiles[userID]
				var text string
				if ok {
					text = fmt.Sprintf("%s | age:%d | gender:%s | location:%s", last.Content, profile.Age, profile.Gender, profile.Location)
				} else {
					text = last.Content
				}
				emb, _, err := getEmbedding(text)
				if err != nil {
					log.Printf("Error generating embedding for user %s: %v", userID, err)
					continue
				}
				insert := InsertRequest{ID: userID, Vector: emb}
				err = postJSON(getVectorDBAPI()+"/insert", insert, nil)
				if err != nil {
					log.Printf("Error updating embedding in vector-db for user %s: %v", userID, err)
				}
				log.Printf("[Job] Embedding updated for user %s", userID)
			}
		}
	}()
}

func main() {
	r := gin.Default()

	r.POST("/interactions", handleInteraction)
	r.GET("/recommendations", handleRecommendations)
	r.POST("/profile", handleProfile)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	startEmbeddingRefresher()

	r.Run(":" + os.Getenv("PORT"))
}

func handleInteraction(c *gin.Context) {
	requestID := genRequestID()
	var inter Interaction
	if err := c.ShouldBindJSON(&inter); err != nil {
		logJSON(map[string]interface{}{"level": "error", "msg": "invalid interaction", "request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "request_id": requestID})
		return
	}

	userInteractions[inter.UserID] = append(userInteractions[inter.UserID], inter)

	emb, elapsed, err := getEmbedding(inter.Content)
	if err != nil {
		logJSON(map[string]interface{}{"level": "error", "msg": "embedding failed", "request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding failed", "request_id": requestID})
		return
	}

	embeddingLatency.Observe(elapsed)

	insert := InsertRequest{ID: inter.UserID, Vector: emb}
	if err := postJSON(getVectorDBAPI()+"/insert", insert, nil); err != nil {
		logJSON(map[string]interface{}{"level": "error", "msg": "vector db insert failed", "request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vector db insert failed", "request_id": requestID})
		return
	}

	logJSON(map[string]interface{}{"level": "info", "msg": "interaction stored", "request_id": requestID, "user_id": inter.UserID})
	c.JSON(http.StatusOK, gin.H{"status": "interaction stored", "request_id": requestID})
}

func handleRecommendations(c *gin.Context) {
	requestID := genRequestID()
	userID := c.Query("user_id")
	if userID == "" {
		logJSON(map[string]interface{}{"level": "error", "msg": "user_id required", "request_id": requestID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required", "request_id": requestID})
		return
	}

	start := time.Now()
	// Search user embedding (example: embedding of the user_id itself)
	emb, elapsed, err := getEmbedding(userID)
	if err != nil {
		logJSON(map[string]interface{}{"level": "error", "msg": "embedding failed", "request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding failed", "request_id": requestID})
		return
	}
	embeddingLatency.Observe(elapsed)

	query := QueryRequest{Vector: emb, K: 5}
	var resp QueryResponse
	if err := postJSON(getVectorDBAPI()+"/query", query, &resp); err != nil {
		logJSON(map[string]interface{}{"level": "error", "msg": "vector db query failed", "request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vector db query failed", "request_id": requestID})
		return
	}
	rankingElapsed := float64(time.Since(start).Milliseconds())
	rankingLatency.Observe(rankingElapsed)

	// Log similarity score
	for i, id := range resp.IDs {
		logJSON(map[string]interface{}{
			"level": "info", "msg": "recommendation", "request_id": requestID, "user_id": userID, "recommended_id": id, "score": resp.Distances[i],
		})
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": resp.IDs, "distances": resp.Distances, "request_id": requestID})
}

func handleProfile(c *gin.Context) {
	var profile UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userProfiles[profile.UserID] = profile
	c.JSON(http.StatusOK, gin.H{"status": "profile updated"})
}

func getEmbedding(text string) ([]float32, float64, error) {
	req := EmbedRequest{Text: text}
	var resp EmbedResponse
	start := time.Now()
	if err := postJSON(getEmbeddingAPI()+"/embed", req, &resp); err != nil {
		return nil, 0, err
	}
	elapsed := float64(time.Since(start).Milliseconds())
	if resp.ElapsedMs > 0 {
		elapsed = resp.ElapsedMs
	}
	return resp.Embedding, elapsed, nil
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

func logJSON(obj map[string]interface{}) {
	b, _ := json.Marshal(obj)
	log.Println(string(b))
}

func genRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.Itoa(rand.Intn(10000))
}

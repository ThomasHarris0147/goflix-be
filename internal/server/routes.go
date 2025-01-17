package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"goflix-be/internal/database"
	"goflix-be/internal/video_processing"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	r.GET("/video", s.VideoRequestHandler)

	r.POST("/video", s.VideoUploadHandler)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

type VideoRequest struct {
	Name    string `json:"name" binding:"required"`
	Quality string `json:"quality" binding:"required"`
}

func (s *Server) VideoRequestHandler(c *gin.Context) {
	var req VideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := make(map[string]string)
	srv := database.New()
	cacheHit, cacheMiss := srv.GetValueRedis(c, req.Name+req.Quality)
	if cacheMiss == nil {
		log.Println("cache hit")
		resp["message"] = cacheHit
	} else {
		result, err := database.GetVideoBasedOnNameAndQuality(req.Name, req.Quality)
		if err != nil {
			resp["message"] = "Error found: " + err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		} else {
			if path, ok := result[0]["path"].(string); ok {
				resp["message"] = path
				srv.SetValueRedis(c, req.Name+req.Quality, path)
			} else {
				log.Fatalf("path is not a string %v", path)
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

type VideoUpload struct {
	Name        string `json:"name" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func LaunchNewWriter() *kafka.Writer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "my-topic",
		Balancer: &kafka.RoundRobin{},
	})
	return writer
}

func SendMessage(writer *kafka.Writer, ctx context.Context, message []byte) {
	err := writer.WriteMessages(ctx,
		kafka.Message{
			Value: message,
		},
	)
	if err != nil {
		log.Fatalf("failed to write message: %v", err)
	}
}

func (s *Server) VideoUploadHandler(c *gin.Context) {
	var req VideoUpload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	writer := LaunchNewWriter()
	defer writer.Close()
	videoPath := req.Path
	w, h := video_processing.GetVideoSize(videoPath)
	fmt.Println("video has been processed, width:", w, "height:", h)
	validCompressionResolutions := video_processing.ReturnValidCompressionRates(h)

	for _, v := range validCompressionResolutions {
		message := &video_processing.VideoCompressionSpec{
			VideoPath:          videoPath,
			CompressionQuality: v,
			ChunkSegments:      10,
			Name:               req.Name,
			Description:        req.Description,
		}
		marshalledData, malshalledErr := json.Marshal(message)
		if malshalledErr != nil {
			panic(malshalledErr)
		}
		SendMessage(writer, c, marshalledData)
	}
	c.JSON(http.StatusOK, "Message Sent!")
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

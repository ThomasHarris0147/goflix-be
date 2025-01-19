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

	r.GET("/clear_redis", s.ClearRedis)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

type VideoRequest struct {
	Name    string `json:"name"`
	Quality string `json:"quality"`
}

type Video struct {
	Id          int    `json:Id`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Path        string `json:"Path"`
	Quality     string `json:"Quality"`
}

func returnAllVideos() map[string]any {
	var result []map[string]interface{}
	var err error
	resp := make(map[string]any)
	result, err = database.GetAllFromTable("videos")
	if err != nil {
		resp["error"] = "Error found: " + err.Error()
		return resp
	}
	resp["data"] = result
	resp["message"] = "successfully retrieved all data"
	return resp
}

func checkRedisCache(c *gin.Context, req VideoRequest, srv database.Service) map[string]any {
	resp := make(map[string]any)
	var cacheHit string
	var cacheMiss error
	cacheHit, cacheMiss = srv.GetValueRedis(c, req.Name+req.Quality)
	if cacheMiss != nil {
		resp["error"] = "missed cache"
		return resp
	} else {
		log.Println("cache hit")
		var video Video
		err := json.Unmarshal([]byte(cacheHit), &video)
		if err != nil {
			resp["error"] = "redis fetch error: " + err.Error()
			return resp
		}
		var result []Video
		result = append(result, video)
		resp["data"] = result
	}
	return resp
}

func validateGetVideoBody(c *gin.Context, req VideoRequest) (map[string]any, VideoRequest) {
	resp := make(map[string]any)
	if err := c.ShouldBindJSON(&req); err != nil {
		resp["error"] = "cannot process body: " + err.Error()
	}
	if req.Quality != "" && req.Name == "" {
		resp["error"] = "If inputting Quality, require Name field aswell"
	}
	if req.Quality == "" && req.Name != "" {
		resp["error"] = "If inputting Name, require Quality field aswell"
	}
	return resp, req
}

func (s *Server) VideoRequestHandler(c *gin.Context) {
	var err error
	contentLength := c.Request.ContentLength
	if contentLength == 0 {
		// Skip processing if there is no body
		c.JSON(http.StatusOK, returnAllVideos())
		return
	}
	var req VideoRequest
	resp, req := validateGetVideoBody(c, req)
	if _, exists := resp["error"]; exists {
		c.JSON(http.StatusBadGateway, resp)
		return
	}
	srv := database.New()
	resp = checkRedisCache(c, req, srv)
	if _, exists := resp["error"]; exists {
		var result []map[string]interface{}
		result, err = database.GetVideoBasedOnNameAndQuality(req.Name, req.Quality)
		if err != nil {
			resp["message"] = "Error found: " + err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		} else {
			log.Println(result[0])
			srv.SetValueRedis(c, req.Name+req.Quality, result[0])
			resp["data"] = result
		}
	}
	resp["message"] = "successfully retrieved all data"
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

func (s *Server) ClearRedis(c *gin.Context) {
	// ONLY HERE FOR TESTING REASONS
	srv := database.New()
	srv.ClearAllValuesRedis(c)
	c.JSON(http.StatusOK, "Redis Cleared!")
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

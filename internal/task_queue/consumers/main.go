package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"goflix-be/internal/database"
	video_processing "goflix-be/internal/video_processing"

	"github.com/segmentio/kafka-go"
)

func launchNewReader() *kafka.Reader {
	log.Println("Starting reader with group id: my-group, topic: my-topic, brokers: localhost:9092, partition: 1")
	ReaderConfig := kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "my-group-1",
		Topic:   "my-topic",
	}
	r := kafka.NewReader(ReaderConfig)
	log.Println("Reader started")
	return r
}

func closeOnSigInterruptOrSigTerm(reader *kafka.Reader) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Println("shutting down reader gracefully:", sig)
			err := reader.Close()
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		}
	}()
}

func get_message(r *kafka.Reader) string {
	msg, err := r.ReadMessage(context.Background())

	if err != nil {
		log.Fatalf("failed to read message: %v", err)
	}

	return string(msg.Value)
}

func main() {
	r := launchNewReader()
	log.Println("Waiting for message...")
	for {
		closeOnSigInterruptOrSigTerm(r)
		msg := get_message(r)

		fmt.Println("Received message: ", msg, "\nnow processing video")
		videoSpecs := &video_processing.VideoCompressionSpec{}
		json.Unmarshal([]byte(msg), &videoSpecs)
		fmt.Println("video has been processed, quality:", videoSpecs.CompressionQuality)
		videoName := strings.Split(videoSpecs.VideoPath, "/")[len(strings.Split(videoSpecs.VideoPath, "/"))-1]
		videoNameWoExt := strings.Split(videoName, ".")[0]
		videoPath := "/Users/thomasharris/side-projects/goflix/backend/goflix-be/test/data/" + videoNameWoExt
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			err := os.Mkdir(videoPath, 0755)
			if err != nil {
				panic(err)
			}
		}
		video_processing.ChangeCodec(videoSpecs.VideoPath, videoPath+"/output_"+videoSpecs.CompressionQuality+".mp4", videoSpecs.CompressionQuality)
		fmt.Println("video has been compressed, quality:", videoSpecs.CompressionQuality)
		fmt.Println("now chunking video")
		inputVideo := []string{videoPath + "/output_" + videoSpecs.CompressionQuality + ".mp4"}
		video_processing.ChunkVideo(inputVideo, videoSpecs.ChunkSegments)
		database.InsertInto("videos",
			[]string{"name", "description", "path", "quality"},
			[]string{videoSpecs.Name, videoSpecs.Description, inputVideo[0], videoSpecs.CompressionQuality})
	}
}

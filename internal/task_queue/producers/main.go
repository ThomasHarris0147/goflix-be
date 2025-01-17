package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goflix-be/internal/video_processing"
	"log"

	"github.com/segmentio/kafka-go"
)

func debugNumOfPartitions() {
	// Connect to Kafka broker
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		log.Fatalf("failed to connect to Kafka: %v", err)
	}
	defer conn.Close()

	topic := "my-topic"

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		log.Fatalf("failed to fetch partitions: %v", err)
	}

	log.Printf("Topic %q has %d partitions", topic, len(partitions))

	for _, p := range partitions {
		log.Printf("Partition: %d, Leader: %d, Replicas: %v, ISR: %v",
			p.ID, p.Leader, p.Replicas, p.Isr)
	}
}

func LaunchNewWriter() *kafka.Writer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "my-topic",
		Balancer: &kafka.RoundRobin{},
	})
	return writer
}

func send_message(writer *kafka.Writer, ctx context.Context, message []byte) {
	err := writer.WriteMessages(ctx,
		kafka.Message{
			Value: message,
		},
	)
	if err != nil {
		log.Fatalf("failed to write message: %v", err)
	}
}

func main() {
	// debugNumOfPartitions()
	ctx := context.Background()
	writer := LaunchNewWriter()
	defer writer.Close()
	exampleVideoPath := "/Users/thomasharris/side-projects/goflix/backend/goflix-be/test/data/Skepta, Flo Milli - Why Lie.mp4"
	w, h := video_processing.GetVideoSize(exampleVideoPath)
	fmt.Println("video has been processed, width:", w, "height:", h)
	validCompressionResolutions := video_processing.ReturnValidCompressionRates(h)

	for _, v := range validCompressionResolutions {
		message := &video_processing.VideoCompressionSpec{
			VideoPath:          exampleVideoPath,
			CompressionQuality: v,
			ChunkSegments:      10,
		}
		marshalledData, malshalledErr := json.Marshal(message)
		if malshalledErr != nil {
			panic(malshalledErr)
		}
		send_message(writer, ctx, marshalledData)
	}
}

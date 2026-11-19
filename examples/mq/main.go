package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type job struct {
	ID      int    `json:"id"`
	Payload string `json:"payload"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "produce":
		err = runProducer(os.Args[2:])
	case "worker":
		err = runWorkerPool(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: go run . <produce|worker> [flags]")
}

func runProducer(args []string) (resultErr error) {
	flags := flag.NewFlagSet("produce", flag.ContinueOnError)
	brokers := flags.String("brokers", envOrDefault("KAFKA_BROKERS", "localhost:9092"), "comma-separated Kafka brokers")
	topic := flags.String("topic", envOrDefault("KAFKA_TOPIC", "jobs"), "Kafka topic")
	count := flags.Int("count", 10, "number of jobs to publish")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *count < 1 {
		return errors.New("count must be greater than zero")
	}
	brokerList := splitBrokers(*brokers)
	if len(brokerList) == 0 {
		return errors.New("at least one Kafka broker is required")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokerList...),
		Topic:        *topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	defer func() {
		if err := writer.Close(); err != nil && resultErr == nil {
			resultErr = fmt.Errorf("close Kafka writer: %w", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for id := 1; id <= *count; id++ {
		body, err := json.Marshal(job{ID: id, Payload: "example work item"})
		if err != nil {
			return fmt.Errorf("encode job %d: %w", id, err)
		}
		if err := writer.WriteMessages(ctx, kafka.Message{Key: []byte(strconv.Itoa(id)), Value: body}); err != nil {
			return fmt.Errorf("publish job %d: %w", id, err)
		}
		log.Printf("published job id=%d", id)
	}
	return nil
}

func runWorkerPool(args []string) error {
	flags := flag.NewFlagSet("worker", flag.ContinueOnError)
	brokers := flags.String("brokers", envOrDefault("KAFKA_BROKERS", "localhost:9092"), "comma-separated Kafka brokers")
	topic := flags.String("topic", envOrDefault("KAFKA_TOPIC", "jobs"), "Kafka topic")
	groupID := flags.String("group", envOrDefault("KAFKA_GROUP", "example-workers"), "Kafka consumer group")
	workers := flags.Int("workers", 3, "number of worker processes in this pool")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *workers < 1 {
		return errors.New("workers must be greater than zero")
	}
	brokerList := splitBrokers(*brokers)
	if len(brokerList) == 0 {
		return errors.New("at least one Kafka broker is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runWorkers(ctx, brokerList, *topic, *groupID, *workers)
}

func runWorkers(ctx context.Context, brokers []string, topic, groupID string, count int) error {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, count)
	for id := 1; id <= count; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := consume(workerCtx, brokers, topic, groupID, id); err != nil && ctx.Err() == nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
			}
		}(id)
	}
	wg.Wait()

	if ctx.Err() != nil {
		return nil
	}
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func consume(ctx context.Context, brokers []string, topic, groupID string, workerID int) (resultErr error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer func() {
		if err := reader.Close(); err != nil && resultErr == nil {
			resultErr = fmt.Errorf("worker %d close reader: %w", workerID, err)
		}
	}()
	log.Printf("worker %d started (group=%s topic=%s)", workerID, groupID, topic)

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("worker %d fetch: %w", workerID, err)
		}

		var item job
		if err := json.Unmarshal(message.Value, &item); err != nil {
			return fmt.Errorf("worker %d decode message at offset %d: %w", workerID, message.Offset, err)
		}
		if err := process(ctx, workerID, item); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("worker %d process job %d: %w", workerID, item.ID, err)
		}
		if err := reader.CommitMessages(ctx, message); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("worker %d commit job %d: %w", workerID, item.ID, err)
		}
		log.Printf("worker %d completed and committed job id=%d", workerID, item.ID)
	}
}

func process(ctx context.Context, workerID int, item job) error {
	log.Printf("worker %d processing job id=%d payload=%q", workerID, item.ID, item.Payload)
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func splitBrokers(value string) []string {
	var brokers []string
	for _, broker := range strings.Split(value, ",") {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

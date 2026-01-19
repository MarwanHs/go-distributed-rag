package main

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/MarwanHs/go-distributed-rag/internal/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	redisClient := setupRedis()
	kafkaWriter := setupKafka()
	defer kafkaWriter.Close()

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	docHandler := handlers.NewDocumentHandler(redisClient, kafkaWriter, logger)

	e.POST("/upload", docHandler.Upload)
	e.GET("/status/:job_id", docHandler.Status)

	e.Logger.Fatal(e.Start(":8080"))
}

func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func setupKafka() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "pdf-processing",
		Balancer: &kafka.LeastBytes{},
	}
}

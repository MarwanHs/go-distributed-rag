package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/MarwanHs/go-distributed-rag/src/gateway/internal/proto/rag/v1"

	"github.com/MarwanHs/go-distributed-rag/src/gateway/internal/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	redisClient := setupRedis()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	logger.Info("Successfully pinged Redis")

	grpcClient, grpcConn := setupGRPCClient()
	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer grpcCancel()
	defer grpcConn.Close()
	resp, err := grpcClient.HealthCheck(grpcCtx, &pb.HealthCheckRequest{})
	if err != nil {
		logger.Warn("Worker health check failed (expected if not running yet)", "error", err)
	} else {
		logger.Info("Worker connection verified", "status", resp.Status)
	}

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

func setupGRPCClient() (pb.RagServiceClient, *grpc.ClientConn) {
	conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := pb.NewRagServiceClient(conn)

	return client, conn
}

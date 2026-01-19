package main

// --- 1. Imports ---
import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/MarwanHs/my-rag-app/internal/job"
)

// --- 2. Global Variables ---
var (
	redisClient *redis.Client
	kafkaWriter *kafka.Writer
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// --- 3. Connect to Redis ---
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// --- 4. Connect to Kafka ---
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "pdf-processing",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	logger.Info("Connected to Infrastructure")

	// --- 5. Setup Echo ---
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

	// Routes
	e.POST("/upload", func(c echo.Context) error {
		return uploadHandler(c, logger)
	})
	e.GET("/status/:job_id", statusHandler)

	logger.Info("API Server starting on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

// --- 6. Upload Handler (The Producer Logic) ---
func uploadHandler(c echo.Context, logger *slog.Logger) error {
	// A. Receive File
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file uploaded"})
	}

	// B. Setup Data
	jobID := fmt.Sprintf("%d", time.Now().UnixNano())
	tempPath := fmt.Sprintf("/tmp/%s", file.Filename)

	job := job.New(jobID, file.Filename, tempPath)

	// C. Update Redis (The Scoreboard) - fire and forget
	ctx := context.Background()
	err = redisClient.Set(ctx, "job:"+jobID, "Pending", 10*time.Minute).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error: " + err.Error(),
		})
	}

	// D. Serialize Job to JSON
	// TODO: Use json.Marshal() to convert your job struct into bytes.
	// This returns ([]byte, error).

	// E. Send to Kafka (The Conveyor Belt)
	// TODO: Use kafkaWriter.WriteMessages(ctx, ...)
	// You need to create a kafka.Message{}.
	//   - Key:   []byte(jobID)  (Ensures order for this specific job)
	//   - Value: The JSON bytes you created in Step D.

	// Hint: If WriteMessages fails, you should probably update Redis to say "Failed"
	// so the user isn't waiting forever.

	// F. Success Response
	return c.JSON(http.StatusAccepted, map[string]string{
		"message": "File accepted for processing",
		"job_id":  jobID,
		"status":  "Pending",
	})
}

// --- 7. Status Handler ---
func statusHandler(c echo.Context) error {
	ctx := c.Request().Context()
	status, err := redisClient.Get(ctx, "job:"+c.Param("job_id")).Result()

	if err == redis.Nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "job not found: " + err.Error(),
		})
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"job_id": c.Param("job_id"),
		"status": status,
	})
}

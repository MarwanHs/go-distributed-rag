package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/MarwanHs/go-distributed-rag/internal/job" // Import your job package
)

type DocumentHandler struct {
	RedisClient *redis.Client
	KafkaWriter *kafka.Writer
	Logger      *slog.Logger
}

func NewDocumentHandler(r *redis.Client, k *kafka.Writer, l *slog.Logger) *DocumentHandler {
	return &DocumentHandler{
		RedisClient: r,
		KafkaWriter: k,
		Logger:      l,
	}
}

func (h *DocumentHandler) Upload(c echo.Context) error {

	file, err := c.FormFile("file")
	if err != nil {
		h.Logger.Error("failed to get file", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not open file"})
	}
	defer src.Close()

	jobID := uuid.New().String()
	tempPath := fmt.Sprintf("/tmp/%s_%s", jobID, file.Filename)

	job := job.New(jobID, file.Filename, tempPath)

	dst, err := os.Create(tempPath)
	if err != nil {
		h.Logger.Error("failed to create temp file", "path", tempPath)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server storage error"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save file"})
	}

	ctx := context.Background()
	err = h.RedisClient.Set(ctx, "job:"+jobID, "Pending", 24*time.Hour).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error: " + err.Error(),
		})
	}

	jobRaw, err := json.Marshal(job)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error: " + err.Error(),
		})
	}

	err = h.KafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(jobID),
		Value: jobRaw,
	})
	if err != nil {
		h.RedisClient.Set(ctx, "job:"+jobID, "Failed", 24*time.Hour)

		h.Logger.Error("failed to write to kafka", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to queue job",
		})
	}

	return c.JSON(http.StatusAccepted, map[string]string{
		"message": "File accepted for processing",
		"job_id":  jobID,
		"status":  "Pending",
	})
}

func (h *DocumentHandler) Status(c echo.Context) error {
	ctx := c.Request().Context()
	status, err := h.RedisClient.Get(ctx, "job:"+c.Param("job_id")).Result()

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

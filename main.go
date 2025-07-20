package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/services"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/storage"
	"github.com/gofiber/fiber/v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := fiber.New()

	atomicCache := cache.NewCostRoutingThresholdCache()
	pg, err := storage.NewPostgresClient(ctx, getPostgresDSN())
	if err != nil {
		panic("erro ao iniciar o banco")
	}

	screening := workers.NewQueueWorker[any](15000)
	highPriority := workers.NewQueueWorker[any](5000)
	lowPriority := workers.NewQueueWorker[any](10000)
	lowWaitingRoom := workers.NewQueueWorker[any](50000)
	highWaitingRoom := workers.NewQueueWorker[any](40000)

	costRoutingThresholdService := services.NewCostRoutingThresholdService(atomicCache, pg)
	screeningService := services.NewScreeningService(atomicCache, highPriority, lowPriority)

	workers.StartWorker(ctx, "threshold", 5*time.Second, costRoutingThresholdService.Calculation)
	workers.StartWorker(ctx, "RetryFallback", 15*time.Second, func(ctx context.Context) error {

		if screening.CountFallback() > 0 {
			screening.RetryFallback()
		}

		if highPriority.CountFallback() > 0 {
			highPriority.RetryFallback()
		}

		if lowPriority.CountFallback() > 0 {
			lowPriority.RetryFallback()
		}

		if lowWaitingRoom.CountFallback() > 0 {
			lowWaitingRoom.RetryFallback()
		}

		if highWaitingRoom.CountFallback() > 0 {
			highWaitingRoom.RetryFallback()
		}

		return nil
	})

	screening.Consume(ctx, 150, screeningService.Redirect)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "API is healthy",
		})
	})

	app.Listen(":8888")
}

func getPostgresDSN() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if port == "" {
		port = "5432"
	}
	println(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname))

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

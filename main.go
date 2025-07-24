package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/services"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
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

	screening := workers.NewQueueWorker(config.Env.ScreeningQueue.Buffer)
	highPriority := workers.NewQueueWorker(config.Env.HighPriorityQueue.Buffer)
	lowPriority := workers.NewQueueWorker(config.Env.LowPriorityQueue.Buffer)
	lowWaitingRoom := workers.NewQueueWorker(config.Env.LowWaitingRoomQueue.Buffer)
	highWaitingRoom := workers.NewQueueWorker(config.Env.HighWaitingRoomQueue.Buffer)

	costRoutingThresholdService := services.NewCostRoutingThresholdService(atomicCache, pg, config.Env.KFactor)
	screeningService := services.NewScreeningService(atomicCache, highPriority, lowPriority)

	workers.StartWorker(ctx, "threshold", 5*time.Second, costRoutingThresholdService.Calculation)
	workers.StartWorker(ctx, "retryFallback", 15*time.Second, func(ctx context.Context) error {

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

	screening.Consume(ctx, config.Env.ScreeningQueue.Workers, screeningService.Redirect)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "API is healthy",
		})
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Env.StartPort))
	if err != nil {
		log.Fatalf("erro ao criar listener: %v", err)
	}

	log.Printf(`

╔════════════════════════════════════════════════════╗
║ 🚀 Servidor Fiber iniciado com sucesso!            ║
║ 📡 Escutando em: http://%s             ║
╚════════════════════════════════════════════════════╝
`, ln.Addr().String())

	if err := app.Listener(ln); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

func getPostgresDSN() string {

	println(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", config.Env.Postgres.User, config.Env.Postgres.Pass, config.Env.Postgres.Host, config.Env.Postgres.PORT, config.Env.Postgres.Name))

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", config.Env.Postgres.User, config.Env.Postgres.Pass, config.Env.Postgres.Host, config.Env.Postgres.PORT, config.Env.Postgres.Name)
}

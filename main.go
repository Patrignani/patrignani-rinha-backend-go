package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/repositories"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/services"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/clients"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/models"
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
	defer pg.Close()

	atomicCache.SetHealthDeafultApi(false)
	atomicCache.SetHealthFallbackApi(false)

	httpClient := clients.NewHttpRequest()

	paymentRepo := repositories.NewPaymentRepository(pg)

	screening := workers.NewQueueWorker(config.Env.ScreeningQueue.Buffer)
	highPriority := workers.NewQueueWorker(config.Env.HighPriorityQueue.Buffer)
	lowPriority := workers.NewQueueWorker(config.Env.LowPriorityQueue.Buffer)
	waitingRoom := workers.NewQueueWorker(config.Env.WaitingRoomQueue.Buffer)

	screeningService := services.NewScreeningService(atomicCache, highPriority, lowPriority, waitingRoom)
	checkHealt := services.NewCheckHealthPaymentService(httpClient, atomicCache)
	waitServer := services.NewWaitingRoomServer(screening)
	paymentServer := services.NewPaymentService(httpClient, waitingRoom, atomicCache, paymentRepo)

	if config.Env.EnableCheckHealthCheck {
		workers.StartWorker(ctx, "healthCheckPayment", 5*time.Second+300*time.Millisecond, checkHealt.SetStatusPayment)
	}

	workers.StartWorker(ctx, "retryFallback", 100*time.Millisecond, func(ctx context.Context) error {

		if screening.CountFallback() > 0 {
			screening.RetryFallback()
		}

		if highPriority.CountFallback() > 0 {
			highPriority.RetryFallback()
		}

		if lowPriority.CountFallback() > 0 {
			lowPriority.RetryFallback()
		}

		if waitingRoom.CountFallback() > 0 {
			waitingRoom.RetryFallback()
		}

		return nil
	})

	go screening.Consume(ctx, config.Env.ScreeningQueue.Workers, screeningService.Redirect)
	go waitingRoom.Consume(ctx, config.Env.WaitingRoomQueue.Workers, waitServer.Delay)
	go highPriority.Consume(ctx, config.Env.HighPriorityQueue.Workers, paymentServer.ExecuteFallback)
	go lowPriority.Consume(ctx, config.Env.LowPriorityQueue.Workers, paymentServer.ExecuteDefault)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "API is healthy",
		})
	})

	app.Post("/payments", func(c *fiber.Ctx) error {
		var payload models.PaymentBasic

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		screening.Send(workers.Message{
			CorrelationId: payload.CorrelationId,
			Amount:        payload.Amount,
			EnqueueAt:     time.Now().UTC(),
		})

		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/payments-summary", func(c *fiber.Ctx) error {
		fromStr := c.Query("from")
		toStr := c.Query("to")

		var fromTime, toTime time.Time
		var err error

		if fromStr != "" {
			fromTime, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid 'from' timestamp format")
			}
		}

		if toStr != "" {
			toTime, err = time.Parse(time.RFC3339, toStr)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid 'to' timestamp format")
			}
		}

		summary, err := paymentRepo.GetPaymentSummary(ctx, &fromTime, &toTime)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "error get summary")
		}

		return c.Status(fiber.StatusOK).JSON(summary)
	})

	app.Delete("/purge", func(c *fiber.Ctx) error {
		ctx := context.Background()
		err := paymentRepo.PurgeAll(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Falha ao deletar dados: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"message": "Todos os dados da tabela entry_history foram deletados.",
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
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", config.Env.Postgres.User, config.Env.Postgres.Pass, config.Env.Postgres.Host, config.Env.Postgres.PORT, config.Env.Postgres.Name)
}

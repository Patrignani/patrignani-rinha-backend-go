package repositories

import (
	"context"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/models"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/storage"
)

type PaymentRepository interface {
	Insert(ctx context.Context, payment models.PaymentDb) error
	GetPaymentSummary(ctx context.Context, from, to *time.Time) (models.SummaryResponse, error)
	PurgeAll(ctx context.Context) error
}

type PaymentRepositoryImp struct {
	pg storage.PostgresClient
}

func NewPaymentRepository(pg storage.PostgresClient) PaymentRepository {
	return &PaymentRepositoryImp{
		pg: pg,
	}
}

func (p *PaymentRepositoryImp) Insert(ctx context.Context, payment models.PaymentDb) error {
	sql := `
		INSERT INTO entry_history (correlationId, amount, fallback, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := p.pg.Exec(ctx, sql,
		payment.CorrelationId,
		payment.Amount,
		payment.Fallback,
		payment.CreatedAt,
	)

	return err
}

func (p *PaymentRepositoryImp) GetPaymentSummary(ctx context.Context, from, to *time.Time) (models.SummaryResponse, error) {
	query := `
		SELECT 
			fallback,
			COUNT(*) AS total_requests,
			SUM(amount) AS total_amount
		FROM 
			entry_history
		WHERE 
			($1::timestamp IS NULL OR created_at >= $1)
			AND ($2::timestamp IS NULL OR created_at <= $2)
		GROUP BY 
			fallback;
	`

	rows, err := p.pg.Query(ctx, query, from, to)
	if err != nil {
		return models.SummaryResponse{}, err
	}
	defer rows.Close()

	var summary models.SummaryResponse

	for rows.Next() {
		var fallback bool
		var totalRequests int
		var totalAmount float64

		if err := rows.Scan(&fallback, &totalRequests, &totalAmount); err != nil {
			return models.SummaryResponse{}, err
		}

		if fallback {
			summary.Fallback = models.PaymentSummary{
				TotalRequests: totalRequests,
				TotalAmount:   totalAmount,
			}
		} else {
			summary.Default = models.PaymentSummary{
				TotalRequests: totalRequests,
				TotalAmount:   totalAmount,
			}
		}
	}

	return summary, nil
}

func (p *PaymentRepositoryImp) PurgeAll(ctx context.Context) error {
	sql := `TRUNCATE TABLE entry_history RESTART IDENTITY;`
	_, err := p.pg.Exec(ctx, sql)
	return err
}

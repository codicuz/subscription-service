package repository

import (
	"context"
	"fmt"
	"github.com/codicuz/subscription-service/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate.ToTime(),
		endDateToTime(sub.EndDate),
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании подписки: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub := &model.Subscription{}
	var startDate, endDate *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении подписки: %w", err)
	}

	if startDate != nil {
		sub.StartDate = model.CustomDate{
			Month: int(startDate.Month()),
			Year:  startDate.Year(),
		}
	}

	if endDate != nil {
		sub.EndDate = &model.CustomDate{
			Month: int(endDate.Month()),
			Year:  endDate.Year(),
		}
	}

	return sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, limit, offset int) ([]*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка подписок: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription

	for rows.Next() {
		sub := &model.Subscription{}
		var startDate, endDate *time.Time

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&startDate,
			&endDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании подписки: %w", err)
		}

		if startDate != nil {
			sub.StartDate = model.CustomDate{
				Month: int(startDate.Month()),
				Year:  startDate.Year(),
			}
		}

		if endDate != nil {
			sub.EndDate = &model.CustomDate{
				Month: int(endDate.Month()),
				Year:  endDate.Year(),
			}
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, id uuid.UUID, input *model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	sub, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}

	if input.ServiceName != nil {
		sub.ServiceName = *input.ServiceName
	}
	if input.Price != nil {
		sub.Price = *input.Price
	}
	if input.StartDate != nil {
		startDate, err := model.ParseDate(*input.StartDate)
		if err != nil {
			return nil, fmt.Errorf("неверный формат start_date: %w", err)
		}
		sub.StartDate = startDate
	}
	if input.EndDate != nil {
		if *input.EndDate == "" {
			sub.EndDate = nil
		} else {
			endDate, err := model.ParseDate(*input.EndDate)
			if err != nil {
				return nil, fmt.Errorf("неверный формат end_date: %w", err)
			}
			sub.EndDate = &endDate
		}
	}

	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err = r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate.ToTime(),
		endDateToTime(sub.EndDate),
		sub.ID,
	).Scan(&sub.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении подписки: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении подписки: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("подписка не найдена")
	}

	return nil
}

func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, req *model.SubscriptionReportRequest) (*model.SubscriptionReport, error) {
	startDate, err := model.ParseDate(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат start_date: %w", err)
	}

	endDate, err := model.ParseDate(req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат end_date: %w", err)
	}

	periodStart := startDate.ToTime()
	periodEnd := model.LastDayOfMonth(endDate)

	query := `
		SELECT COALESCE(SUM(price), 0) as total_cost
		FROM subscriptions
		WHERE start_date <= $1
		AND (end_date IS NULL OR end_date >= $2)
	`

	args := []interface{}{periodEnd, periodStart}

	argIdx := 3
	if req.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, req.UserID)
		argIdx++
	}

	if req.ServiceName != "" {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argIdx)
		args = append(args, "%"+req.ServiceName+"%")
	}

	var totalCost int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подсчете стоимости: %w", err)
	}

	return &model.SubscriptionReport{TotalCost: totalCost}, nil
}

func endDateToTime(endDate *model.CustomDate) *time.Time {
	if endDate == nil {
		return nil
	}
	t := time.Date(endDate.Year, time.Month(endDate.Month), 1, 0, 0, 0, 0, time.UTC)
	return &t
}

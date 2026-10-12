package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CallRecord represents a persisted routing decision.
type CallRecord struct {
	CallID        string
	Caller        string
	Destination   string
	Region        string
	SelectedRoute string
	RoutingStatus string
}

// CallRecordRepository handles call history persistence.
type CallRecordRepository struct {
	DB *pgxpool.Pool
}

func NewCallRecordRepository(db *pgxpool.Pool) *CallRecordRepository {
	return &CallRecordRepository{DB: db}
}

// Save inserts one routing attempt into PostgreSQL.
func (repo *CallRecordRepository) Save(
	ctx context.Context,
	record CallRecord,
) error {

	const query = `
		INSERT INTO call_records (
			call_id,
			caller,
			destination,
			region,
			selected_route,
			routing_status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := repo.DB.Exec(
		dbCtx,
		query,
		record.CallID,
		record.Caller,
		record.Destination,
		record.Region,
		record.SelectedRoute,
		record.RoutingStatus,
	)

	if err != nil {
		return fmt.Errorf("failed to save call record: %w", err)
	}

	return nil
}

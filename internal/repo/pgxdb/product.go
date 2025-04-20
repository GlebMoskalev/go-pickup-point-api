package pgxdb

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, receptionID, productType string) (*entity.Product, error) {
	log := slog.With("layer", "ProductRepo", "operation", "Create", "receptionID", receptionID, "type", productType)
	log.Debug("starting product creation")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	query := `
	INSERT INTO products (type, reception_id) 
	VALUES ($1, $2)
	RETURNING id, date_time, reception_id
`
	var id, receptionUUID uuid.UUID
	var dateTime time.Time

	err = tx.QueryRow(ctx, query, productType, receptionID).Scan(&id, &dateTime, &receptionUUID)
	if err != nil {
		log.Error("failed to create product", "error", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return nil, err
	}

	product := entity.Product{
		ID:          id,
		DateTime:    dateTime,
		Type:        productType,
		ReceptionID: receptionUUID,
	}

	log.Info("product created successfully", "productID", id.String())
	return &product, nil
}

func (r *ProductRepo) DeleteLastProduct(ctx context.Context, receptionID string) error {
	log := slog.With("layer", "ProductRepo", "operation", "DeleteLastProduct", "receptionID", receptionID)
	log.Debug("starting product deletion")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	query := `
	DELETE FROM products
	WHERE id = (
	    SELECT id
	    FROM products
	    WHERE reception_id = $1
	    ORDER BY order_number DESC 
	    LIMIT 1
	)
	RETURNING id
`

	var id uuid.UUID
	err = tx.QueryRow(ctx, query, receptionID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("not found product")
			return repoerr.ErrNoRows
		}
		log.Error("failed to delete product", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return err
	}

	log.Info("product deleted successfully", "productID", id.String())
	return nil
}

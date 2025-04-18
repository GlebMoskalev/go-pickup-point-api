package pgxdb

import (
	"context"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type PVZRepo struct {
	db *pgxpool.Pool
}

func NewPVZRepo(db *pgxpool.Pool) *PVZRepo {
	return &PVZRepo{db: db}
}

func (r *PVZRepo) Create(ctx context.Context, city string) (*entity.PVZ, error) {
	log := slog.With("layer", "PVZRepo", "operation", "Create", "city", city)
	log.Debug("starting pvz creation")

	query := `
	INSERT INTO pvz (city)
	VALUES ($1)
	RETURNING id, registration_date
`
	var id uuid.UUID
	var date time.Time
	err := r.db.QueryRow(ctx, query, city).Scan(&id, &date)
	if err != nil {
		log.Error("failed to create pvz", "error", err)
		return nil, err
	}

	pvz := entity.PVZ{
		ID:               id,
		RegistrationDate: date,
		City:             city,
	}

	log.Info("pvz created successfully", "pvzID", id.String())
	return &pvz, nil
}

func (r *PVZRepo) Exists(ctx context.Context, pvzID string) bool {
	log := slog.With("layer", "PVZRepo", "operation", "Exists", "pvzID", pvzID)
	log.Debug("checking pvz existence")

	query := `
	SELECT EXISTS(
		SELECT 1
		FROM pvz
		WHERE id = $1
	)
`

	var exists bool
	err := r.db.QueryRow(ctx, query, pvzID).Scan(&exists)
	if err != nil {
		log.Error("failed to check pvz existence", "error", err)
		return false
	}

	log.Debug("pvz existence check completed", "exists", exists)
	return exists
}

func (r *PVZRepo) ListWithDetails(ctx context.Context,
	startDate, endDate *time.Time,
	page, limit int) ([]entity.PVZWithDetails, error) {
	log := slog.With("layer", "PVZRepo", "operation", "ListWithDetails", "page", page, "limit", limit)
	log.Debug("starting list pvz with details")

	query := `
	SELECT
	    p.id AS pvz_id, p.registration_date, p.city,
	    r.id AS reception_id, r.date_time AS reception_date_time, r.pvz_id, r.status,
	    pr.id AS product_id, pr.date_time AS product_date_time, pr.type AS product_type
	FROM pvz p
	LEFT JOIN receptions r ON p.id = r.pvz_id
	LEFT JOIN products pr ON r.id = pr.reception_id
`

	var args []any
	var conditions []string
	if startDate != nil {
		conditions = append(conditions, "r.date_time >= $1")
		args = append(args, *startDate)
	}
	if endDate != nil {
		idx := len(args) + 1
		conditions = append(conditions, "r.date_time<=$"+strconv.Itoa(idx))
		args = append(args, *endDate)
	}

	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ")
	}

	idx := len(args) + 1
	query += fmt.Sprintf(`
	ORDER BY p.id, r.date_time DESC, pr.order_number DESC
	LIMIT $%d OFFSET  $%d
`, idx, idx+1)

	args = append(args, limit, (page-1)*limit)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Error("failed to execute query", "error", err)
		return nil, err
	}
	defer rows.Close()

	pvzMap := make(map[string]*entity.PVZWithDetails)
	for rows.Next() {
		var (
			pvzID            uuid.UUID
			registrationDate time.Time
			city             string

			receptionID    pgtype.UUID
			receptionDate  pgtype.Timestamp
			receptionPVZID uuid.UUID
			status         pgtype.Text

			productID   pgtype.UUID
			productDate pgtype.Timestamp
			productType pgtype.Text
		)

		err := rows.Scan(
			&pvzID, &registrationDate, &city,
			&receptionID, &receptionDate, &receptionPVZID, &status,
			&productID, &productDate, &productType,
		)
		if err != nil {
			log.Error("failed to scan row", "error", err)
			return nil, err
		}
		pvz, exists := pvzMap[pvzID.String()]
		if !exists {
			pvz = &entity.PVZWithDetails{
				PVZ: entity.PVZ{
					ID:               pvzID,
					RegistrationDate: registrationDate,
					City:             city,
				},
				Receptions: []entity.ReceptionDetails{},
			}
			pvzMap[pvzID.String()] = pvz
		}

		if receptionID.Valid {
			var reception *entity.ReceptionDetails
			receptionUUID, err := uuid.FromBytes(receptionID.Bytes[:])
			if err != nil {
				log.Error("failed to parse reception id to UUID", "error", err)
				return nil, err
			}

			for i, v := range pvz.Receptions {
				if v.Reception.ID == receptionUUID {
					reception = &pvz.Receptions[i]
					break
				}
			}

			if reception == nil {
				reception = &entity.ReceptionDetails{
					Reception: entity.Reception{
						ID:       receptionUUID,
						DateTime: receptionDate.Time,
						PVZID:    receptionPVZID,
						Status:   status.String,
					},
					Products: []entity.Product{},
				}
				pvz.Receptions = append(pvz.Receptions, *reception)
			}

			if productID.Valid {
				productUUID, err := uuid.FromBytes(productID.Bytes[:])
				if err != nil {
					log.Error("failed to parse product id to UUID", "error", err)
					return nil, err
				}

				product := entity.Product{
					ID:          productUUID,
					DateTime:    productDate.Time,
					Type:        productType.String,
					ReceptionID: receptionUUID,
				}

				reception.Products = append(reception.Products, product)
			}
		}
	}

	if err := rows.Err(); err != nil {
		log.Error("rows error", "error", err)
		return nil, err
	}
	result := make([]entity.PVZWithDetails, 0, len(pvzMap))

	for _, pvz := range pvzMap {
		result = append(result, *pvz)
	}

	log.Info("pvz list retrieved successfully", "count", len(result))
	return result, nil
}

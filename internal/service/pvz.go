package service

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/metrics"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"log/slog"
	"time"
)

type PVZService struct {
	pvzRepo repo.PVZ
}

func NewPVZService(pvzRepo repo.PVZ) *PVZService {
	return &PVZService{pvzRepo: pvzRepo}
}

func (s *PVZService) Create(ctx context.Context, city string) (*entity.PVZ, error) {
	log := slog.With("layer", "PVZService", "operation", "Create", "city", city)
	log.Debug("starting create pvz")

	if city != entity.CityMoscow && city != entity.CityKazan && city != entity.CitySPB {
		return nil, ErrInvalidCity
	}

	pvz, err := s.pvzRepo.Create(ctx, city)
	if err != nil {
		log.Error("failed to create pvz", "error", err)
		return nil, ErrInternal
	}

	metrics.PVZCreated.Inc()
	log.Info("pvz created successfully", "pvzID", pvz.ID.String())
	return pvz, nil
}

func (s *PVZService) ListWithDetails(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.PVZWithDetails, error) {
	log := slog.With("layer", "PVZService", "operation", "ListWithDetails", "page", page, "limit", limit)
	log.Debug("starting list pvz with details")

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 30 {
		limit = 30
	}

	pvzs, err := s.pvzRepo.ListWithDetails(ctx, startDate, endDate, page, limit)
	if err != nil {
		log.Error("failed to get list pvz with details", "error", err)
		return nil, ErrInternal
	}

	log.Info("pvz list with details get successfully")
	return pvzs, err
}

package service

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"log/slog"
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

	log.Info("pvz created successfully", "pvzID", pvz.ID.String())
	return pvz, nil
}

package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"log/slog"
)

type ReceptionService struct {
	receptionRepo repo.Reception
	pvzRepo       repo.PVZ
}

func NewReceptionService(receptionRepo repo.Reception, pvzRepo repo.PVZ) *ReceptionService {
	return &ReceptionService{receptionRepo: receptionRepo, pvzRepo: pvzRepo}
}

func (s *ReceptionService) Create(ctx context.Context, pvzID string) (*entity.Reception, error) {
	log := slog.With("layer", "ReceptionService", "operation", "Create", "pvzID", pvzID)
	log.Debug("starting reception creation")

	if !s.pvzRepo.Exists(ctx, pvzID) {
		log.Error("pvz does not exist")
		return nil, ErrInvalidPVZID
	}

	hasOpen, err := s.receptionRepo.HasOpenReception(ctx, pvzID)
	if err != nil {
		log.Error("failed to check open reception", "error", err)
		return nil, ErrInternal
	}

	if hasOpen {
		log.Error("open reception already exists")
		return nil, ErrOpenReceptionExists
	}

	reception, err := s.receptionRepo.Create(ctx, pvzID)
	if err != nil {
		log.Error("failed to create reception", "error", err)
		return nil, ErrInternal
	}
	log.Info("reception created successfully", "receptionID", reception.ID.String())
	return reception, nil
}

func (s *ReceptionService) CloseLastReception(ctx context.Context, pvzID string) error {
	log := slog.With("layer", "ReceptionService", "operation", "CloseLastReception", "pvzID", pvzID)
	log.Debug("starting reception closure")

	if !s.pvzRepo.Exists(ctx, pvzID) {
		log.Error("pvz does not exist")
		return ErrInvalidPVZID
	}

	reception, err := s.receptionRepo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		if errors.Is(err, repoerr.ErrNoRows) {
			log.Error("not found open reception")
			return ErrNoOpenReception
		}
		log.Error("failed to check open reception", "error", err)
		return ErrInternal
	}

	err = s.receptionRepo.Close(ctx, reception.ID.String())
	if err != nil {
		if errors.Is(err, repoerr.ErrNoRows) {
			log.Error("not found open reception")
			return ErrNoOpenReception
		}
		log.Error("failed to close reception", "error", err)
		return ErrInternal
	}

	log.Info("reception closed successfully")
	return nil
}

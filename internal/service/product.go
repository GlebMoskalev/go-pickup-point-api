package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/metrics"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"log/slog"
)

type ProductService struct {
	productRepo   repo.Product
	receptionRepo repo.Reception
	pvzRepo       repo.PVZ
}

func NewProductService(productRepo repo.Product, receptionRepo repo.Reception, pvzRepo repo.PVZ) *ProductService {
	return &ProductService{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, pvzID, productType string) (*entity.Product, error) {
	log := slog.With("layer", "ProductService", "operation", "Create", "pvzID", pvzID, "type", productType)
	log.Debug("starting product creation")

	if productType != entity.ProductTypeClothes &&
		productType != entity.ProductTypeElectronics &&
		productType != entity.ProductTypeShoes {
		return nil, ErrInvalidProductType
	}

	if !s.pvzRepo.Exists(ctx, pvzID) {
		log.Error("pvz does not exist")
		return nil, ErrInvalidPVZID
	}

	reception, err := s.receptionRepo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		if errors.Is(err, repoerr.ErrNoRows) {
			log.Error("not found open error")
			return nil, ErrNoOpenReception
		}
		log.Error("failed to check open reception", "error", err)
		return nil, ErrInternal
	}

	product, err := s.productRepo.Create(ctx, reception.ID.String(), productType)
	if err != nil {
		log.Error("failed to create product", "error", err)
		return nil, ErrInternal
	}

	metrics.ProductsAdded.Inc()
	log.Info("product created successfully", "productID", product.ID.String(), "receptionID", reception.ID.String())
	return product, nil
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID string) error {
	log := slog.With("layer", "ProductService", "operation", "DeleteLastProduct", "pvzID", pvzID)
	log.Debug("starting product deletion")

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

	log = log.With("receptionID", reception.ID.String())

	err = s.productRepo.DeleteLastProduct(ctx, reception.ID.String())
	if err != nil {
		if errors.Is(err, repoerr.ErrNoRows) {
			log.Error("not found product")
			return ErrNoProducts
		}
		log.Error("failed to delete product", "error", err)
		return err
	}

	log.Info("product deleted successfully")
	return nil
}

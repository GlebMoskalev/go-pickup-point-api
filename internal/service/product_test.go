package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/mocks"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestProductService_Create(t *testing.T) {
	testCases := []struct {
		name            string
		pvzID           string
		productType     string
		prepareRepos    func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ)
		expectedProduct *entity.Product
		expectedError   error
	}{
		{
			name:        "successful creation",
			pvzID:       uuid.New().String(),
			productType: entity.ProductTypeElectronics,
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{ID: receptionID, Status: "in_progress"}, nil)
				productID := uuid.New()
				productRepo.On("Create", mock.Anything, receptionID.String(), entity.ProductTypeElectronics).
					Return(&entity.Product{
						ID:          productID,
						DateTime:    time.Now(),
						Type:        entity.ProductTypeElectronics,
						ReceptionID: receptionID,
					}, nil)
			},
			expectedProduct: &entity.Product{
				ID:          uuid.UUID{},
				DateTime:    time.Time{},
				Type:        entity.ProductTypeElectronics,
				ReceptionID: uuid.UUID{},
			},
			expectedError: nil,
		},
		{
			name:            "invalid product type",
			pvzID:           uuid.New().String(),
			productType:     "invalid",
			prepareRepos:    func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {},
			expectedProduct: nil,
			expectedError:   ErrInvalidProductType,
		},
		{
			name:        "invalid pvz id",
			pvzID:       uuid.New().String(),
			productType: entity.ProductTypeClothes,
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(false)
			},
			expectedProduct: nil,
			expectedError:   ErrInvalidPVZID,
		},
		{
			name:        "no open reception",
			pvzID:       uuid.New().String(),
			productType: entity.ProductTypeShoes,
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, repoerr.ErrNoRows)
			},
			expectedProduct: nil,
			expectedError:   ErrNoOpenReception,
		},
		{
			name:        "reception repo error",
			pvzID:       uuid.New().String(),
			productType: entity.ProductTypeElectronics,
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("database error"))
			},
			expectedProduct: nil,
			expectedError:   ErrInternal,
		},
		{
			name:        "product repo error",
			pvzID:       uuid.New().String(),
			productType: entity.ProductTypeClothes,
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{ID: receptionID, Status: "in_progress"}, nil)
				productRepo.On("Create", mock.Anything, receptionID.String(), entity.ProductTypeClothes).
					Return(nil, errors.New("database error"))
			},
			expectedProduct: nil,
			expectedError:   ErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productRepo := mocks.NewProduct(t)
			receptionRepo := mocks.NewReception(t)
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepos(productRepo, receptionRepo, pvzRepo)

			service := NewProductService(productRepo, receptionRepo, pvzRepo)
			ctx := context.Background()

			product, err := service.Create(ctx, tc.pvzID, tc.productType)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				_, err := uuid.Parse(product.ID.String())
				assert.NoError(t, err, "Product ID should be a valid UUID")
				_, err = uuid.Parse(product.ReceptionID.String())
				assert.NoError(t, err, "ReceptionID should be a valid UUID")
				assert.Equal(t, tc.expectedProduct.Type, product.Type)
			}
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	databaseErr := errors.New("database error")

	testCases := []struct {
		name          string
		pvzID         string
		prepareRepos  func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ)
		expectedError error
	}{
		{
			name:  "successful deletion",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{ID: receptionID, Status: "in_progress"}, nil)
				productRepo.On("DeleteLastProduct", mock.Anything, receptionID.String()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "invalid pvz id",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(false)
			},
			expectedError: ErrInvalidPVZID,
		},
		{
			name:  "no open reception",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, repoerr.ErrNoRows)
			},
			expectedError: ErrNoOpenReception,
		},
		{
			name:  "no products",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{ID: receptionID, Status: "in_progress"}, nil)
				productRepo.On("DeleteLastProduct", mock.Anything, receptionID.String()).
					Return(repoerr.ErrNoRows)
			},
			expectedError: ErrNoProducts,
		},
		{
			name:  "reception repo error",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("database error"))
			},
			expectedError: ErrInternal,
		},
		{
			name:  "product repo error",
			pvzID: uuid.New().String(),
			prepareRepos: func(productRepo *mocks.Product, receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{ID: receptionID, Status: "open"}, nil)
				productRepo.On("DeleteLastProduct", mock.Anything, receptionID.String()).
					Return(databaseErr)
			},
			expectedError: databaseErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productRepo := mocks.NewProduct(t)
			receptionRepo := mocks.NewReception(t)
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepos(productRepo, receptionRepo, pvzRepo)

			service := NewProductService(productRepo, receptionRepo, pvzRepo)

			err := service.DeleteLastProduct(context.Background(), tc.pvzID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

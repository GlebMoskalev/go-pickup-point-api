package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestPVZService_Create(t *testing.T) {
	testCases := []struct {
		name          string
		city          string
		prepareRepo   func(repo *mocks.PVZ)
		expectedPVZ   *entity.PVZ
		expectedError error
	}{
		{
			name: "successful creation",
			city: entity.CityMoscow,
			prepareRepo: func(repo *mocks.PVZ) {
				pvzID := uuid.New()
				repo.On("Create", mock.Anything, entity.CityMoscow).
					Return(&entity.PVZ{
						ID:               pvzID,
						RegistrationDate: time.Now(),
						City:             entity.CityMoscow,
					}, nil)
			},
			expectedPVZ: &entity.PVZ{
				ID:               uuid.UUID{},
				RegistrationDate: time.Time{},
				City:             entity.CityMoscow,
			},
			expectedError: nil,
		},
		{
			name:          "invalid city",
			city:          "InvalidCity",
			prepareRepo:   func(repo *mocks.PVZ) {},
			expectedPVZ:   nil,
			expectedError: ErrInvalidCity,
		},
		{
			name: "repo error",
			city: entity.CityKazan,
			prepareRepo: func(repo *mocks.PVZ) {
				repo.On("Create", mock.Anything, entity.CityKazan).
					Return(nil, errors.New("database error"))
			},
			expectedPVZ:   nil,
			expectedError: ErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepo(pvzRepo)
			service := NewPVZService(pvzRepo)
			ctx := context.Background()

			pvz, err := service.Create(ctx, tc.city)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, pvz)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				_, err := uuid.Parse(pvz.ID.String())
				assert.NoError(t, err, "PVZ ID should be a valid UUID")
				assert.Equal(t, tc.expectedPVZ.City, pvz.City)
				assert.False(t, pvz.RegistrationDate.IsZero(), "RegistrationDate should be set")
			}
		})
	}
}

func TestPVZService_ListWithDetails(t *testing.T) {
	startDate := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 4, 18, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		name          string
		startDate     *time.Time
		endDate       *time.Time
		page          int
		limit         int
		prepareRepo   func(repo *mocks.PVZ)
		expectedPVZs  []entity.PVZWithDetails
		expectedError error
	}{
		{
			name:      "successful list with details",
			startDate: &startDate,
			endDate:   &endDate,
			page:      1,
			limit:     10,
			prepareRepo: func(repo *mocks.PVZ) {
				pvzID := uuid.New()
				receptionID := uuid.New()
				productID := uuid.New()
				repo.On("ListWithDetails", mock.Anything, &startDate, &endDate, 1, 10).
					Return([]entity.PVZWithDetails{
						{
							PVZ: entity.PVZ{
								ID:               pvzID,
								RegistrationDate: time.Now(),
								City:             entity.CityMoscow,
							},
							Receptions: []entity.ReceptionDetails{
								{
									Reception: entity.Reception{
										ID:       receptionID,
										DateTime: time.Now(),
										PVZID:    pvzID,
										Status:   "in_progress",
									},
									Products: []entity.Product{
										{
											ID:          productID,
											DateTime:    time.Now(),
											Type:        entity.ProductTypeElectronics,
											ReceptionID: receptionID,
										},
									},
								},
							},
						},
					}, nil)
			},
			expectedPVZs: []entity.PVZWithDetails{
				{
					PVZ: entity.PVZ{
						ID:               uuid.UUID{},
						RegistrationDate: time.Time{},
						City:             entity.CityMoscow,
					},
					Receptions: []entity.ReceptionDetails{
						{
							Reception: entity.Reception{
								ID:       uuid.UUID{},
								DateTime: time.Time{},
								PVZID:    uuid.UUID{},
								Status:   "in_progress",
							},
							Products: []entity.Product{
								{
									ID:          uuid.UUID{},
									DateTime:    time.Time{},
									Type:        entity.ProductTypeElectronics,
									ReceptionID: uuid.UUID{},
								},
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:      "invalid page adjusted",
			startDate: nil,
			endDate:   nil,
			page:      0,
			limit:     10,
			prepareRepo: func(repo *mocks.PVZ) {
				repo.On("ListWithDetails", mock.Anything, (*time.Time)(nil), (*time.Time)(nil), 1, 10).
					Return([]entity.PVZWithDetails{}, nil)
			},
			expectedPVZs:  []entity.PVZWithDetails{},
			expectedError: nil,
		},
		{
			name:      "invalid limit adjusted",
			startDate: nil,
			endDate:   nil,
			page:      1,
			limit:     50,
			prepareRepo: func(repo *mocks.PVZ) {
				repo.On("ListWithDetails", mock.Anything, (*time.Time)(nil), (*time.Time)(nil), 1, 30).
					Return([]entity.PVZWithDetails{}, nil)
			},
			expectedPVZs:  []entity.PVZWithDetails{},
			expectedError: nil,
		},
		{
			name:      "repo error",
			startDate: &startDate,
			endDate:   &endDate,
			page:      1,
			limit:     10,
			prepareRepo: func(repo *mocks.PVZ) {
				repo.On("ListWithDetails", mock.Anything, &startDate, &endDate, 1, 10).
					Return(nil, errors.New("database error"))
			},
			expectedPVZs:  nil,
			expectedError: ErrInternal,
		},
		{
			name:      "no date filters",
			startDate: nil,
			endDate:   nil,
			page:      1,
			limit:     10,
			prepareRepo: func(repo *mocks.PVZ) {
				repo.On("ListWithDetails", mock.Anything, (*time.Time)(nil), (*time.Time)(nil), 1, 10).
					Return([]entity.PVZWithDetails{}, nil)
			},
			expectedPVZs:  []entity.PVZWithDetails{},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepo(pvzRepo)
			service := NewPVZService(pvzRepo)
			ctx := context.Background()

			pvzs, err := service.ListWithDetails(ctx, tc.startDate, tc.endDate, tc.page, tc.limit)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, pvzs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedPVZs), len(pvzs))
				for i, pvz := range pvzs {
					_, err := uuid.Parse(pvz.PVZ.ID.String())
					assert.NoError(t, err, "PVZ ID should be a valid UUID")
					assert.False(t, pvz.PVZ.RegistrationDate.IsZero(), "RegistrationDate should be set")
					if len(tc.expectedPVZs) > i {
						assert.Equal(t, tc.expectedPVZs[i].PVZ.City, pvz.PVZ.City)
						assert.Equal(t, len(tc.expectedPVZs[i].Receptions), len(pvz.Receptions))
						for j, reception := range pvz.Receptions {
							_, err := uuid.Parse(reception.Reception.ID.String())
							assert.NoError(t, err, "Reception ID should be a valid UUID")
							_, err = uuid.Parse(reception.Reception.PVZID.String())
							assert.NoError(t, err, "Reception PVZID should be a valid UUID")
							assert.False(t, reception.Reception.DateTime.IsZero(), "Reception DateTime should be set")
							if len(tc.expectedPVZs[i].Receptions) > j {
								assert.Equal(t, tc.expectedPVZs[i].Receptions[j].Reception.Status, reception.Reception.Status)
								assert.Equal(t, len(tc.expectedPVZs[i].Receptions[j].Products), len(reception.Products))
								for k, product := range reception.Products {
									_, err := uuid.Parse(product.ID.String())
									assert.NoError(t, err, "Product ID should be a valid UUID")
									_, err = uuid.Parse(product.ReceptionID.String())
									assert.NoError(t, err, "Product ReceptionID should be a valid UUID")
									assert.False(t, product.DateTime.IsZero(), "Product DateTime should be set")
									if len(tc.expectedPVZs[i].Receptions[j].Products) > k {
										assert.Equal(t, tc.expectedPVZs[i].Receptions[j].Products[k].Type, product.Type)
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

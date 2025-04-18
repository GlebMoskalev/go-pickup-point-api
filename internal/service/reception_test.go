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

func TestReceptionService_Create(t *testing.T) {
	testCases := []struct {
		name              string
		pvzID             string
		prepareRepos      func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ)
		expectedReception *entity.Reception
		expectedError     error
	}{
		{
			name:  "successful creation",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("HasOpenReception", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				receptionID := uuid.New()
				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{
						ID:       receptionID,
						DateTime: time.Now(),
						PVZID:    uuid.UUID{},
						Status:   "in_progress",
					}, nil)
			},
			expectedReception: &entity.Reception{
				ID:       uuid.UUID{},
				DateTime: time.Time{},
				PVZID:    uuid.UUID{},
				Status:   "in_progress",
			},
			expectedError: nil,
		},
		{
			name:  "invalid pvz id",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(false)
			},
			expectedReception: nil,
			expectedError:     ErrInvalidPVZID,
		},
		{
			name:  "open reception exists",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("HasOpenReception", mock.Anything, mock.AnythingOfType("string")).Return(true, nil)
			},
			expectedReception: nil,
			expectedError:     ErrOpenReceptionExists,
		},
		{
			name:  "has open reception error",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("HasOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(false, errors.New("database error"))
			},
			expectedReception: nil,
			expectedError:     ErrInternal,
		},
		{
			name:  "create reception error",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("HasOpenReception", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("database error"))
			},
			expectedReception: nil,
			expectedError:     ErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receptionRepo := mocks.NewReception(t)
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepos(receptionRepo, pvzRepo)

			service := NewReceptionService(receptionRepo, pvzRepo)
			ctx := context.Background()

			reception, err := service.Create(ctx, tc.pvzID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, reception)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reception)
				_, err := uuid.Parse(reception.ID.String())
				assert.NoError(t, err, "Reception ID should be a valid UUID")
				_, err = uuid.Parse(reception.PVZID.String())
				assert.NoError(t, err, "PVZID should be a valid UUID")
				assert.Equal(t, tc.expectedReception.Status, reception.Status)
				assert.False(t, reception.DateTime.IsZero(), "DateTime should be set")
			}
		})
	}
}

func TestReceptionService_CloseLastReception(t *testing.T) {
	testCases := []struct {
		name          string
		pvzID         string
		prepareRepos  func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ)
		expectedError error
	}{
		{
			name:  "successful closure",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{
						ID:       receptionID,
						DateTime: time.Now(),
						PVZID:    uuid.UUID{},
						Status:   "in_progress",
					}, nil)
				receptionRepo.On("Close", mock.Anything, receptionID.String()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "invalid pvz id",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(false)
			},
			expectedError: ErrInvalidPVZID,
		},
		{
			name:  "no open reception",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, repoerr.ErrNoRows)
			},
			expectedError: ErrNoOpenReception,
		},
		{
			name:  "get open reception error",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("database error"))
			},
			expectedError: ErrInternal,
		},
		{
			name:  "close reception error",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{
						ID:       receptionID,
						DateTime: time.Now(),
						PVZID:    uuid.UUID{},
						Status:   "in_progress",
					}, nil)
				receptionRepo.On("Close", mock.Anything, receptionID.String()).
					Return(errors.New("database error"))
			},
			expectedError: ErrInternal,
		},
		{
			name:  "close reception not found",
			pvzID: uuid.New().String(),
			prepareRepos: func(receptionRepo *mocks.Reception, pvzRepo *mocks.PVZ) {
				pvzRepo.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(true)
				receptionID := uuid.New()
				receptionRepo.On("GetLastOpenReception", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{
						ID:       receptionID,
						DateTime: time.Now(),
						PVZID:    uuid.UUID{},
						Status:   "in_progress",
					}, nil)
				receptionRepo.On("Close", mock.Anything, receptionID.String()).
					Return(repoerr.ErrNoRows)
			},
			expectedError: ErrNoOpenReception,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receptionRepo := mocks.NewReception(t)
			pvzRepo := mocks.NewPVZ(t)
			tc.prepareRepos(receptionRepo, pvzRepo)

			service := NewReceptionService(receptionRepo, pvzRepo)
			ctx := context.Background()

			err := service.CloseLastReception(ctx, tc.pvzID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

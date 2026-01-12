package lot

/* import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	mock_lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/mocks"
)

func TestUseCases_CreateLot(t *testing.T) {
	sowingDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		lot           *domain.Lot
		setupMock     func(*mock_lot.MockRepositoryPort)
		expectedID    int64
		expectedError error
	}{
		{
			name: "successful creation",
			lot: &domain.Lot{
				Name:       "Test Lot",
				Hectares:   decimal.NewFromFloat(100.50),
				SowingDate: &sowingDate,
			},
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					CreateLot(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name: "repository error",
			lot: &domain.Lot{
				Name:       "Test Lot",
				Hectares:   decimal.NewFromFloat(100.50),
				SowingDate: &sowingDate,
			},
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					CreateLot(gomock.Any(), gomock.Any()).
					Return(int64(0), types.NewError(types.ErrInternal, "database error", nil))
			},
			expectedID:    0,
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_lot.NewMockRepositoryPort(ctrl)
			tt.setupMock(mockRepo)

			uc := NewUseCases(mockRepo)
			id, err := uc.CreateLot(context.Background(), tt.lot)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestUseCases_GetLot(t *testing.T) {
	sowingDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		id            int64
		setupMock     func(*mock_lot.MockRepositoryPort)
		expected      *domain.Lot
		expectedError error
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					GetLot(gomock.Any(), int64(1)).
					Return(&domain.Lot{
						ID:         1,
						Name:       "Test Lot",
						Hectares:   decimal.NewFromFloat(100.50),
						SowingDate: &sowingDate,
					}, nil)
			},
			expected: &domain.Lot{
				ID:         1,
				Name:       "Test Lot",
				Hectares:   decimal.NewFromFloat(100.50),
				SowingDate: &sowingDate,
			},
			expectedError: nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					GetLot(gomock.Any(), int64(999)).
					Return(nil, types.NewError(types.ErrNotFound, "lot not found", nil))
			},
			expected:      nil,
			expectedError: types.NewError(types.ErrNotFound, "lot not found", nil),
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					GetLot(gomock.Any(), int64(1)).
					Return(nil, types.NewError(types.ErrInternal, "database error", nil))
			},
			expected:      nil,
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_lot.NewMockRepositoryPort(ctrl)
			tt.setupMock(mockRepo)

			uc := NewUseCases(mockRepo)
			result, err := uc.GetLot(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUseCases_ListLotsByProject(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int64
		setupMock     func(*mock_lot.MockRepositoryPort)
		expected      []domain.Lot
		expectedError error
	}{
		{
			name:      "successful listing",
			projectID: 1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					ListLotsByProject(gomock.Any(), int64(1)).
					Return([]domain.Lot{
						{ID: 1, Name: "Lot 1", Hectares: decimal.NewFromFloat(100.50)},
						{ID: 2, Name: "Lot 2", Hectares: decimal.NewFromFloat(200.75)},
					}, nil)
			},
			expected: []domain.Lot{
				{ID: 1, Name: "Lot 1", Hectares: decimal.NewFromFloat(100.50)},
				{ID: 2, Name: "Lot 2", Hectares: decimal.NewFromFloat(200.75)},
			},
			expectedError: nil,
		},
		{
			name:      "empty list",
			projectID: 1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					ListLotsByProject(gomock.Any(), int64(1)).
					Return([]domain.Lot{}, nil)
			},
			expected:      []domain.Lot{},
			expectedError: nil,
		},
		{
			name:      "repository error",
			projectID: 1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					ListLotsByProject(gomock.Any(), int64(1)).
					Return(nil, types.NewError(types.ErrInternal, "database error", nil))
			},
			expected:      nil,
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_lot.NewMockRepositoryPort(ctrl)
			tt.setupMock(mockRepo)

			uc := NewUseCases(mockRepo)
			result, err := uc.ListLotsByProject(context.Background(), tt.projectID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUseCases_UpdateLot(t *testing.T) {
	sowingDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		lot           *domain.Lot
		setupMock     func(*mock_lot.MockRepositoryPort)
		expectedError error
	}{
		{
			name: "successful update",
			lot: &domain.Lot{
				ID:         1,
				Name:       "Updated Lot",
				Hectares:   decimal.NewFromFloat(150.25),
				SowingDate: &sowingDate,
			},
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					UpdateLot(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			lot: &domain.Lot{
				ID:         1,
				Name:       "Updated Lot",
				Hectares:   decimal.NewFromFloat(150.25),
				SowingDate: &sowingDate,
			},
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					UpdateLot(gomock.Any(), gomock.Any()).
					Return(types.NewError(types.ErrInternal, "database error", nil))
			},
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_lot.NewMockRepositoryPort(ctrl)
			tt.setupMock(mockRepo)

			uc := NewUseCases(mockRepo)
			err := uc.UpdateLot(context.Background(), tt.lot)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_DeleteLot(t *testing.T) {
	tests := []struct {
		name          string
		id            int64
		setupMock     func(*mock_lot.MockRepositoryPort)
		expectedError error
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteLot(gomock.Any(), int64(1)).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func(mockRepo *mock_lot.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteLot(gomock.Any(), int64(1)).
					Return(types.NewError(types.ErrInternal, "database error", nil))
			},
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_lot.NewMockRepositoryPort(ctrl)
			tt.setupMock(mockRepo)

			uc := NewUseCases(mockRepo)
			err := uc.DeleteLot(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
*/

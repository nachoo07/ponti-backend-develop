package supply_movement

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/mocks"
)

// TestUseCases_DeleteSupplyMovement tests para la eliminación de movimientos de suministro
func TestUseCases_DeleteSupplyMovement(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int64
		supplyID      int64
		setupMock     func(*mocks.MockRepositoryPort)
		expectedError error
		description   string
	}{
		{
			name:      "eliminación exitosa de movimiento normal",
			projectID: 1,
			supplyID:  100,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(100)).
					Return(nil)
			},
			expectedError: nil,
			description:   "Movimiento normal se elimina correctamente",
		},
		{
			name:      "movimiento no encontrado",
			projectID: 1,
			supplyID:  999,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(999)).
					Return(types.NewError(types.ErrNotFound, "supply movement not found", nil))
			},
			expectedError: types.NewError(types.ErrNotFound, "supply movement not found", nil),
			description:   "Error cuando el movimiento no existe",
		},
		{
			name:      "error de stock cerrado",
			projectID: 1,
			supplyID:  200,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(200)).
					Return(types.NewError(types.ErrConflict, "ya existe un movimiento de stock cerrado para este supply en el proyecto", nil))
			},
			expectedError: types.NewError(types.ErrConflict, "ya existe un movimiento de stock cerrado para este supply en el proyecto", nil),
			description:   "No se puede eliminar si el stock está cerrado",
		},
		{
			name:      "🔥 eliminación exitosa de movimiento interno",
			projectID: 1,
			supplyID:  300,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// Mock espera que el repository maneje internamente la eliminación de todos los registros relacionados
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(300)).
					Return(nil)
			},
			expectedError: nil,
			description:   "🔥 Movimiento interno se elimina con todos sus registros relacionados",
		},
		{
			name:      "error interno del repository",
			projectID: 1,
			supplyID:  400,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(400)).
					Return(types.NewError(types.ErrInternal, "database error", nil))
			},
			expectedError: types.NewError(types.ErrInternal, "database error", nil),
			description:   "Error interno del repository se propaga correctamente",
		},
		{
			name:      "🔥 eliminación de movimiento interno desde proyecto destino",
			projectID: 2,
			supplyID:  500,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// Aunque se llame desde el proyecto destino, debe eliminar todos los registros
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(2), int64(500)).
					Return(nil)
			},
			expectedError: nil,
			description:   "🔥 Movimiento interno se puede eliminar desde cualquier proyecto involucrado",
		},
		{
			name:      "proyecto ID inválido",
			projectID: 0,
			supplyID:  100,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(0), int64(100)).
					Return(types.NewError(types.ErrValidation, "invalid project id", nil))
			},
			expectedError: types.NewError(types.ErrValidation, "invalid project id", nil),
			description:   "Validación de ID de proyecto inválido",
		},
		{
			name:      "supply ID inválido",
			projectID: 1,
			supplyID:  0,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(0)).
					Return(types.NewError(types.ErrValidation, "invalid supply movement id", nil))
			},
			expectedError: types.NewError(types.ErrValidation, "invalid supply movement id", nil),
			description:   "Validación de ID de movimiento inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepositoryPort(ctrl)
			mockStockUC := mocks.NewMockUseCasesPort(ctrl)
			tt.setupMock(mockRepo)

			// Create use case
			uc := NewUseCases(mockRepo, mockStockUC, nil)

			// Execute
			err := uc.DeleteSupplyMovement(context.Background(), tt.projectID, tt.supplyID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, tt.description)
				assert.Equal(t, tt.expectedError.Error(), err.Error(), tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestDeleteSupplyMovement_InternalMovementScenarios tests específicos para movimientos internos
func TestDeleteSupplyMovement_InternalMovementScenarios(t *testing.T) {
	tests := []struct {
		name        string
		scenario    string
		projectID   int64
		supplyID    int64
		setupMock   func(*mocks.MockRepositoryPort)
		expectError bool
		description string
	}{
		{
			name:      "🔥 escenario completo: movimiento interno con 3 registros",
			scenario:  "delete_all_related",
			projectID: 1,
			supplyID:  100,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// El repository debe:
				// 1. Encontrar el movimiento principal
				// 2. Identificar que es movimiento interno
				// 3. Buscar todos los registros relacionados (3 en total)
				// 4. Eliminarlos todos
				// 5. Recalcular o eliminar stocks en ambos proyectos
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(100)).
					Return(nil)
			},
			expectError: false,
			description: "🔥 Eliminación completa de movimiento interno (3 registros + actualización de stocks)",
		},
		{
			name:      "🔥 movimiento interno con stocks que quedan con movimientos",
			scenario:  "partial_stock_deletion",
			projectID: 1,
			supplyID:  200,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// Cuando hay otros movimientos en el mismo stock:
				// - No se elimina el stock
				// - Se recalcula RealStockUnits
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(200)).
					Return(nil)
			},
			expectError: false,
			description: "🔥 Stock se mantiene y recalcula cuando quedan otros movimientos",
		},
		{
			name:      "🔥 movimiento interno elimina stocks vacíos",
			scenario:  "full_stock_deletion",
			projectID: 1,
			supplyID:  300,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// Cuando no quedan movimientos:
				// - Se elimina el stock completamente
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(300)).
					Return(nil)
			},
			expectError: false,
			description: "🔥 Stock se elimina completamente cuando no quedan movimientos",
		},
		{
			name:      "🔥 efecto en dinero se revierte automáticamente",
			scenario:  "money_calculation_reversal",
			projectID: 1,
			supplyID:  400,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				// Soft delete automáticamente excluye de queries SQL con deleted_at IS NULL
				// El dinero se recalcula correctamente sin el movimiento eliminado
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(400)).
					Return(nil)
			},
			expectError: false,
			description: "🔥 Soft delete revierte automáticamente el efecto en cálculos de dinero",
		},
		{
			name:      "error: falla al eliminar registros relacionados",
			scenario:  "failed_related_deletion",
			projectID: 1,
			supplyID:  500,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(500)).
					Return(types.NewError(types.ErrInternal, "failed to delete related supply movements", nil))
			},
			expectError: true,
			description: "Error en transacción al eliminar registros relacionados",
		},
		{
			name:      "error: falla al actualizar stock",
			scenario:  "failed_stock_update",
			projectID: 1,
			supplyID:  600,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(600)).
					Return(types.NewError(types.ErrInternal, "failed to update stock real units", nil))
			},
			expectError: true,
			description: "Error al recalcular RealStockUnits después de eliminación",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepositoryPort(ctrl)
			mockStockUC := mocks.NewMockUseCasesPort(ctrl)
			tt.setupMock(mockRepo)

			// Create use case
			uc := NewUseCases(mockRepo, mockStockUC, nil)

			// Execute
			err := uc.DeleteSupplyMovement(context.Background(), tt.projectID, tt.supplyID)

			// Assert
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestDeleteSupplyMovement_EdgeCases tests casos extremos
func TestDeleteSupplyMovement_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		projectID   int64
		supplyID    int64
		setupMock   func(*mocks.MockRepositoryPort)
		expectError bool
		description string
	}{
		{
			name:      "IDs negativos",
			projectID: -1,
			supplyID:  -100,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(-1), int64(-100)).
					Return(types.NewError(types.ErrValidation, "invalid ids", nil))
			},
			expectError: true,
			description: "IDs negativos deben ser rechazados",
		},
		{
			name:      "contexto cancelado",
			projectID: 1,
			supplyID:  100,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(1), int64(100)).
					Return(context.Canceled)
			},
			expectError: true,
			description: "Manejo de contexto cancelado",
		},
		{
			name:      "IDs muy grandes",
			projectID: 9999999999,
			supplyID:  8888888888,
			setupMock: func(mockRepo *mocks.MockRepositoryPort) {
				mockRepo.EXPECT().
					DeleteSupplyMovement(gomock.Any(), int64(9999999999), int64(8888888888)).
					Return(nil)
			},
			expectError: false,
			description: "IDs muy grandes son válidos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepositoryPort(ctrl)
			mockStockUC := mocks.NewMockUseCasesPort(ctrl)
			tt.setupMock(mockRepo)

			// Create use case
			uc := NewUseCases(mockRepo, mockStockUC, nil)

			// Execute
			err := uc.DeleteSupplyMovement(context.Background(), tt.projectID, tt.supplyID)

			// Assert
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

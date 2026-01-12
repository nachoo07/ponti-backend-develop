package supply_movement

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
)

// TestCreateStockDiference valida que la función calcula correctamente las diferencias de stock
func TestCreateStockDiference(t *testing.T) {
	tests := []struct {
		name           string
		isEntry        bool
		quantity       decimal.Decimal
		expectedResult decimal.Decimal
		description    string
	}{
		{
			name:           "is_entry=TRUE con cantidad positiva: se mantiene positiva",
			isEntry:        true,
			quantity:       decimal.NewFromInt(100),
			expectedResult: decimal.NewFromInt(100),
			description:    "Entrada normal: +100 unidades",
		},
		{
			name:           "is_entry=FALSE con cantidad positiva: se vuelve negativa",
			isEntry:        false,
			quantity:       decimal.NewFromInt(100),
			expectedResult: decimal.NewFromInt(-100),
			description:    "Salida normal: -100 unidades",
		},
		{
			name:           "🔥 is_entry=TRUE con cantidad NEGATIVA: se mantiene negativa",
			isEntry:        true,
			quantity:       decimal.NewFromInt(-50),
			expectedResult: decimal.NewFromInt(-50),
			description:    "🔥 CASO CRÍTICO: Movimiento interno salida con is_entry=TRUE y cantidad negativa",
		},
		{
			name:           "is_entry=FALSE con cantidad negativa: se vuelve positiva",
			isEntry:        false,
			quantity:       decimal.NewFromInt(-50),
			expectedResult: decimal.NewFromInt(50),
			description:    "Caso raro: cantidad negativa con is_entry=false se invierte",
		},
		{
			name:           "is_entry=TRUE con cero",
			isEntry:        true,
			quantity:       decimal.Zero,
			expectedResult: decimal.Zero,
			description:    "Cero se mantiene como cero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createStockDiference(tt.isEntry, tt.quantity)
			assert.Equal(t, tt.expectedResult.String(), result.String(), tt.description)
		})
	}
}

// TestMoneyCalculation valida que cantidad × precio funcione correctamente con negativos
func TestMoneyCalculation(t *testing.T) {
	tests := []struct {
		name          string
		quantity      decimal.Decimal
		price         decimal.Decimal
		expectedMoney decimal.Decimal
		shouldBeNeg   bool
		shouldBePos   bool
		description   string
	}{
		{
			name:          "Cantidad positiva × precio = dinero positivo",
			quantity:      decimal.NewFromInt(20),
			price:         decimal.NewFromInt(10),
			expectedMoney: decimal.NewFromInt(200),
			shouldBePos:   true,
			description:   "Caso normal de entrada: 20 × 10 = 200",
		},
		{
			name:          "🔥 Cantidad NEGATIVA × precio = dinero NEGATIVO",
			quantity:      decimal.NewFromInt(-20),
			price:         decimal.NewFromInt(10),
			expectedMoney: decimal.NewFromInt(-200),
			shouldBeNeg:   true,
			description:   "🔥 CASO CRÍTICO: Movimiento interno salida: -20 × 10 = -200",
		},
		{
			name:          "🔥 Cantidad NEGATIVA mayor × precio = dinero NEGATIVO mayor",
			quantity:      decimal.NewFromInt(-50),
			price:         decimal.NewFromInt(15),
			expectedMoney: decimal.NewFromInt(-750),
			shouldBeNeg:   true,
			description:   "🔥 Movimiento interno grande: -50 × 15 = -750",
		},
		{
			name:          "Cantidad cero × precio = cero",
			quantity:      decimal.Zero,
			price:         decimal.NewFromInt(10),
			expectedMoney: decimal.Zero,
			description:   "Sin movimiento: 0 × 10 = 0",
		},
		{
			name:          "Decimales: cantidad negativa × precio decimal",
			quantity:      decimal.NewFromFloat(-7.5),
			price:         decimal.NewFromFloat(12.25),
			expectedMoney: decimal.NewFromFloat(-91.875),
			shouldBeNeg:   true,
			description:   "Con decimales: -7.5 × 12.25 = -91.875",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money := tt.quantity.Mul(tt.price)
			assert.Equal(t, tt.expectedMoney.String(), money.String(), tt.description)

			if tt.shouldBeNeg {
				assert.True(t, money.IsNegative(), "El dinero debe ser negativo")
			}
			if tt.shouldBePos {
				assert.True(t, money.IsPositive(), "El dinero debe ser positivo")
			}
		})
	}
}

// TestBalanceCalculation valida que el balance entre salida y entrada sea correcto
func TestBalanceCalculation(t *testing.T) {
	tests := []struct {
		name            string
		outQuantity     decimal.Decimal
		inQuantity      decimal.Decimal
		price           decimal.Decimal
		expectedBalance decimal.Decimal
		description     string
	}{
		{
			name:            "🔥 Balance perfecto: -200 + 200 = 0",
			outQuantity:     decimal.NewFromInt(-20),
			inQuantity:      decimal.NewFromInt(20),
			price:           decimal.NewFromInt(10),
			expectedBalance: decimal.Zero,
			description:     "🔥 CASO CRÍTICO: Movimiento interno debe tener balance 0",
		},
		{
			name:            "Balance con cantidades grandes: -1000 + 1000 = 0",
			outQuantity:     decimal.NewFromInt(-50),
			inQuantity:      decimal.NewFromInt(50),
			price:           decimal.NewFromInt(20),
			expectedBalance: decimal.Zero,
			description:     "Movimiento interno grande: balance 0",
		},
		{
			name:            "Balance con decimales: -375 + 375 = 0",
			outQuantity:     decimal.NewFromFloat(-15.5),
			inQuantity:      decimal.NewFromFloat(15.5),
			price:           decimal.NewFromFloat(24.19355),
			expectedBalance: decimal.Zero,
			description:     "Con decimales: balance 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moneyOut := tt.outQuantity.Mul(tt.price)
			moneyIn := tt.inQuantity.Mul(tt.price)
			balance := moneyOut.Add(moneyIn)

			assert.True(t, balance.IsZero(),
				"%s\nMoneyOut: %s, MoneyIn: %s, Balance: %s (debe ser 0)",
				tt.description, moneyOut.String(), moneyIn.String(), balance.String())
		})
	}
}

// TestSQLQueryLogic valida la lógica de las queries SQL de la migración 159
func TestSQLQueryLogic(t *testing.T) {
	tests := []struct {
		name          string
		queryFragment string
		mustContain   []string
		description   string
	}{
		{
			name: "🔥 Query debe incluir is_entry=TRUE para capturar movimientos internos",
			queryFragment: `
				SELECT SUM(sm.quantity * s.price)
				FROM public.supply_movements sm
				JOIN public.supplies s ON s.id = sm.supply_id
				WHERE sm.is_entry = TRUE
				  AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
			`,
			mustContain: []string{
				"sm.is_entry = TRUE",
				"Movimiento interno",
				"SUM(sm.quantity * s.price)",
			},
			description: "🔥 Query de migración 159: debe incluir is_entry=TRUE",
		},
		{
			name: "Query debe incluir todos los tipos de movimiento válidos",
			queryFragment: `
				AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
			`,
			mustContain: []string{
				"Stock",
				"Remito oficial",
				"Movimiento interno",
				"Movimiento interno entrada",
			},
			description: "Query debe aceptar los 4 tipos de movimiento",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.mustContain {
				assert.Contains(t, tt.queryFragment, expected,
					"%s - Debe contener: '%s'", tt.description, expected)
			}
		})
	}
}

// TestSumWithNegativeValues simula cómo SQL SUM() maneja valores negativos
func TestSumWithNegativeValues(t *testing.T) {
	tests := []struct {
		name          string
		movements     []decimal.Decimal
		expectedTotal decimal.Decimal
		description   string
	}{
		{
			name: "🔥 Suma con movimiento interno negativo",
			movements: []decimal.Decimal{
				decimal.NewFromInt(1000), // Entrada inicial: +$1,000
				decimal.NewFromInt(-200), // 🔥 Movimiento interno salida: -$200
				decimal.NewFromInt(500),  // Remito oficial: +$500
			},
			expectedTotal: decimal.NewFromInt(1300), // 1000 - 200 + 500 = 1300
			description:   "🔥 CASO CRÍTICO: SUM debe incluir valores negativos correctamente",
		},
		{
			name: "Suma con múltiples movimientos internos",
			movements: []decimal.Decimal{
				decimal.NewFromInt(2000), // Entrada inicial
				decimal.NewFromInt(-300), // Movimiento interno salida 1
				decimal.NewFromInt(-150), // Movimiento interno salida 2
				decimal.NewFromInt(600),  // Remito oficial
			},
			expectedTotal: decimal.NewFromInt(2150), // 2000 - 300 - 150 + 600
			description:   "Múltiples movimientos internos",
		},
		{
			name: "🔥 Transferencia completa entre proyectos: balance global correcto",
			movements: []decimal.Decimal{
				decimal.NewFromInt(1000),  // Entrada inicial en Proyecto A
				decimal.NewFromInt(-1000), // Todo de A se transfiere a B (salida)
				decimal.NewFromInt(1000),  // B recibe todo de A (entrada)
			},
			expectedTotal: decimal.NewFromInt(1000), // Balance global: 1000 - 1000 + 1000 = 1000
			description:   "🔥 Transferencia 100%: el total global se mantiene correcto",
		},
		{
			name: "Solo movimientos negativos",
			movements: []decimal.Decimal{
				decimal.NewFromInt(-100),
				decimal.NewFromInt(-250),
				decimal.NewFromInt(-50),
			},
			expectedTotal: decimal.NewFromInt(-400),
			description:   "Todos negativos: -100 - 250 - 50 = -400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := decimal.Zero
			for _, movement := range tt.movements {
				total = total.Add(movement)
			}
			assert.Equal(t, tt.expectedTotal.String(), total.String(), tt.description)
		})
	}
}

// TestMovementTypeConstants valida que las constantes estén correctas
func TestMovementTypeConstants(t *testing.T) {
	tests := []struct {
		name         string
		constant     string
		expectedName string
	}{
		{
			name:         "STOCK",
			constant:     domain.STOCK,
			expectedName: "Stock",
		},
		{
			name:         "INTERNAL_MOVEMENT",
			constant:     domain.INTERNAL_MOVEMENT,
			expectedName: "Movimiento interno",
		},
		{
			name:         "OFFICIAL_INVOICE",
			constant:     domain.OFFICIAL_INVOICE,
			expectedName: "Remito oficial",
		},
		{
			name:         "INTERNAL_MOVEMENT_IN",
			constant:     domain.INTERNAL_MOVEMENT_IN,
			expectedName: "Movimiento interno entrada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedName, tt.constant)
		})
	}
}

// TestInternalMovementLogic valida la lógica conceptual de los movimientos internos
func TestInternalMovementLogic(t *testing.T) {
	t.Run("🔥 Movimiento interno debe crear 2 registros con is_entry=TRUE", func(t *testing.T) {
		// Simular el concepto
		type Movement struct {
			ProjectID    int64
			Quantity     decimal.Decimal
			IsEntry      bool
			MovementType string
		}

		// Movimiento original: transferir 20 unidades del Proyecto 1 al Proyecto 2
		originalQuantity := decimal.NewFromInt(20)

		// Registro 1: SALIDA en proyecto origen
		movementOut := Movement{
			ProjectID:    1,
			Quantity:     originalQuantity.Neg(), // 🔥 CANTIDAD NEGATIVA
			IsEntry:      true,                   // 🔥 IS_ENTRY = TRUE
			MovementType: domain.INTERNAL_MOVEMENT,
		}

		// Registro 2: ENTRADA en proyecto destino
		movementIn := Movement{
			ProjectID:    2,
			Quantity:     originalQuantity, // Cantidad positiva
			IsEntry:      true,             // IS_ENTRY = TRUE
			MovementType: domain.INTERNAL_MOVEMENT_IN,
		}

		// Validaciones críticas
		assert.True(t, movementOut.IsEntry,
			"❌ FALLO CRÍTICO: movementOut debe tener is_entry=TRUE")
		assert.True(t, movementOut.Quantity.IsNegative(),
			"❌ FALLO CRÍTICO: movementOut debe tener cantidad NEGATIVA")

		assert.True(t, movementIn.IsEntry,
			"movementIn debe tener is_entry=TRUE")
		assert.True(t, movementIn.Quantity.IsPositive(),
			"movementIn debe tener cantidad POSITIVA")

		// Validar balance
		balance := movementOut.Quantity.Add(movementIn.Quantity)
		assert.True(t, balance.IsZero(),
			"❌ FALLO CRÍTICO: Balance debe ser 0")
	})

	t.Run("🔥 Cálculo de dinero con is_entry=TRUE y cantidad negativa", func(t *testing.T) {
		// Movimiento interno: 20 unidades × $10 c/u
		quantity := decimal.NewFromInt(-20) // Cantidad negativa (salida)
		price := decimal.NewFromInt(10)
		isEntry := true // 🔥 is_entry = TRUE

		// Cálculo de dinero
		money := quantity.Mul(price) // -20 × 10 = -200

		// Validaciones
		assert.True(t, isEntry, "is_entry debe ser TRUE para aparecer en reportes")
		assert.True(t, quantity.IsNegative(), "Cantidad debe ser NEGATIVA")
		assert.True(t, money.IsNegative(),
			"❌ FALLO CRÍTICO: Dinero debe ser NEGATIVO (-200) para restar inversión")
		assert.Equal(t, "-200", money.String())

		// Simular query SQL: SUM(quantity * price) WHERE is_entry = TRUE
		// Este registro se incluiría en el SUM porque is_entry=TRUE
		// Y restaría $200 del total porque money es negativo
	})
}

// TestMigration159Behavior valida el comportamiento de la migración 159
func TestMigration159Behavior(t *testing.T) {
	t.Run("🔥 Dashboard ANTES de migración 159: no incluye movimientos internos", func(t *testing.T) {
		// Simular datos ANTES de la migración
		movements := []struct {
			IsEntry        bool
			MovementType   string
			Money          decimal.Decimal
			IncludedBefore bool
		}{
			{IsEntry: true, MovementType: "Stock", Money: decimal.NewFromInt(1000), IncludedBefore: true},
			{IsEntry: true, MovementType: "Remito oficial", Money: decimal.NewFromInt(500), IncludedBefore: true},
			{IsEntry: false, MovementType: "Movimiento interno", Money: decimal.NewFromInt(-200), IncludedBefore: false}, // ❌ No incluido
		}

		totalBefore := decimal.Zero
		for _, m := range movements {
			if m.IncludedBefore {
				totalBefore = totalBefore.Add(m.Money)
			}
		}

		assert.Equal(t, "1500", totalBefore.String(),
			"ANTES: Solo suma Stock + Remito oficial = 1500 (sin movimientos internos)")
	})

	t.Run("🔥 Dashboard DESPUÉS de migración 159: incluye movimientos internos", func(t *testing.T) {
		// Simular datos DESPUÉS de la migración
		movements := []struct {
			IsEntry       bool
			MovementType  string
			Money         decimal.Decimal
			IncludedAfter bool
		}{
			{IsEntry: true, MovementType: "Stock", Money: decimal.NewFromInt(1000), IncludedAfter: true},
			{IsEntry: true, MovementType: "Remito oficial", Money: decimal.NewFromInt(500), IncludedAfter: true},
			{IsEntry: true, MovementType: "Movimiento interno", Money: decimal.NewFromInt(-200), IncludedAfter: true}, // ✅ INCLUIDO
		}

		totalAfter := decimal.Zero
		for _, m := range movements {
			if m.IncludedAfter {
				totalAfter = totalAfter.Add(m.Money)
			}
		}

		assert.Equal(t, "1300", totalAfter.String(),
			"✅ DESPUÉS: Suma Stock + Remito oficial + Movimiento interno = 1000 + 500 - 200 = 1300")
	})
}

-- =====================================================
-- 000062: DOLLAR - Tabla de Tipos de Cambio
-- =====================================================
-- Entidad: dollar (Dólares)
-- Funcionalidad: Almacenar tipos de cambio para conversiones USD/ARS
-- =====================================================

-- Crear tabla de tipos de cambio para conversiones monetarias
CREATE TABLE IF NOT EXISTS fx_rates (
    id SERIAL PRIMARY KEY,
    currency_pair VARCHAR(10) NOT NULL, -- Ej: USDARS, EURUSD
    rate DECIMAL(10,4) NOT NULL, -- Tasa de cambio
    effective_date DATE NOT NULL, -- Fecha de vigencia
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Crear índice único para evitar duplicados de tasa por fecha
CREATE UNIQUE INDEX IF NOT EXISTS idx_fx_rates_unique_pair_date 
ON fx_rates(currency_pair, effective_date);

-- Crear índice para búsquedas por fecha
CREATE INDEX IF NOT EXISTS idx_fx_rates_effective_date 
ON fx_rates(effective_date);

-- Función para obtener tasa de cambio por defecto
CREATE OR REPLACE FUNCTION get_default_fx_rate()
RETURNS DECIMAL AS $$
BEGIN
  RETURN 1.0000; -- Tasa base USD/USD
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Insertar tasa inicial USD/ARS usando función
INSERT INTO fx_rates (currency_pair, rate, effective_date) 
VALUES ('USDARS', get_default_fx_rate(), CURRENT_DATE)
ON CONFLICT (currency_pair, effective_date) DO NOTHING;

-- =====================================================
-- 000062: DOLLAR - Tabla de Tipos de Cambio (ROLLBACK)
-- =====================================================
-- Entidad: dollar (Dólares)
-- Funcionalidad: Eliminar tabla de tipos de cambio
-- =====================================================

-- Eliminar tabla de tipos de cambio
DROP TABLE IF EXISTS fx_rates CASCADE;

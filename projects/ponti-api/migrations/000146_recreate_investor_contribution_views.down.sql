-- ========================================
-- MIGRATION 000146: RECREATE INVESTOR CONTRIBUTION VIEWS (DOWN/ROLLBACK)
-- ========================================
-- 
-- Purpose: Revertir recreación de vistas de investor contribution
-- Date: 2025-10-14
-- Author: System
--
-- Note: No eliminamos las vistas en el rollback porque la migración 135
--       las debe haber creado. Solo las dejamos como estaban.

BEGIN;

-- No hay acción necesaria en el rollback
-- Las vistas vuelven al estado de la migración 135

COMMIT;


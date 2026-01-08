-- ========================================
-- MIGRACIÓN 000181: ROLLBACK - Controles 6, 7, 9 (DOWN)
-- ========================================
--
-- Propósito: Revertir consolidación de fixes de controles 6, 7, 9
-- Fecha: 2024-11-04
--

BEGIN;

-- Esta migración consolidada no tiene rollback específico
-- ya que revierte a un estado previo a las migraciones 000175-000179
-- que fueron aplicadas manualmente.
--
-- Para revertir, se deben aplicar los archivos .down.sql originales
-- en orden inverso: 000179, 000178, 000177, 000176, 000175

RAISE NOTICE 'Para revertir esta migración, aplicar manualmente los .down.sql de 000175-000179 en orden inverso';

COMMIT;


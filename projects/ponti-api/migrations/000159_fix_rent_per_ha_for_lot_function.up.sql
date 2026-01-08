-- ========================================
-- MIGRATION 000159: FIX RENT PER HA FOR LOT FUNCTION (UP)
-- ========================================
-- 
-- Purpose: Corregir función v3_calc.rent_per_ha_for_lot() que fue rota en migración 000094
-- Date: 2025-10-21
-- Author: System
-- 
-- Problema identificado:
--   La migración 000094 sobrescribió v3_calc.rent_per_ha_for_lot() con una versión simplificada
--   que solo devuelve lease_type_value, perdiendo toda la lógica de cálculo por tipo de arriendo.
--
-- Impacto actual:
--   - Proyectos con lease_type_id=1 (% sobre ingreso) muestran rent=0 en informes
--   - Esto afecta "Arriendo / Ha" en informe de Aportes por Inversor
--
-- Solución:
--   Restaurar la función a su implementación correcta que considera todos los tipos de arriendo:
--   - Tipo 1: % sobre ingreso neto
--   - Tipo 2: % sobre (ingreso - costos - admin)
--   - Tipo 3: Valor fijo
--   - Tipo 4: Valor fijo + % sobre ingreso
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- RESTAURAR FUNCIÓN CORRECTA: v3_calc.rent_per_ha_for_lot()
-- ============================================================================

CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Restaura la lógica original que considera todos los tipos de arriendo
  SELECT v3_calc.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_calc.rent_per_ha_for_lot(bigint) IS 
  'SSOT: Calcula el arriendo por hectárea para un lote según el tipo de arriendo configurado. Corregido en migración 159 para restaurar lógica completa.';


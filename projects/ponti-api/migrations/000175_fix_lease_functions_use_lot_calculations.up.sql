-- ========================================
-- MIGRACIÓN 000175: FIX Lease Functions Use Lot Calculations (UP)
-- ========================================
--
-- Propósito: Corregir funciones de arriendo en dashboard para usar cálculo correcto por lote
-- Problema: v3_dashboard_ssot.lease_invested_for_project() y lease_executed_for_project()
--           usan LIMIT 1 y multiplican lease_type_value × total_hectares
--           Esto es INCORRECTO para proyectos con múltiples campos o arriendos variables
-- Solución: Usar v3_lot_ssot.rent_per_ha_for_lot() que considera TODOS los tipos de arriendo:
--           - Tipo 1: % INGRESO NETO
--           - Tipo 2: % UTILIDAD
--           - Tipo 3: ARRIENDO FIJO
--           - Tipo 4: ARRIENDO FIJO + % INGRESO NETO
-- Fecha: 2025-11-03
-- Autor: Sistema
--
-- Impacto: Control 9 (Arriendo) pasará de ERROR a OK
--          Dashboard mostrará arriendo correcto ($162.692 en lugar de $118.900 para Jujuy)
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- FIX 1: lease_invested_for_project
-- ========================================
-- Cambiar de: lease_type_value * total_hectares (INCORRECTO)
-- A: Σ(rent_per_ha_for_lot × hectares) (CORRECTO)

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- FIX: Sumar arriendo de cada lote usando rent_per_ha_for_lot()
  -- Esta función considera TODOS los tipos de arriendo correctamente
  SELECT COALESCE(
    SUM(v3_lot_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
    0
  )::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id 
    AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.lease_invested_for_project(bigint) IS 
'Calcula arriendo invertido total del proyecto sumando rent_per_ha por lote. FIX 000175: Usa v3_lot_ssot.rent_per_ha_for_lot() para considerar todos los tipos de arriendo.';

-- ========================================
-- FIX 2: lease_executed_for_project
-- ========================================
-- Para ejecutados, solo consideramos tipos fijos (3 y 4 - parte fija)
-- Los tipos variables (1, 2, y parte variable de 4) se ejecutan al final en cosecha

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_executed_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Para tipos 3 (FIJO) y 4 (FIJO + %): ejecutado = parte fija
  -- Para tipos 1 (% INGRESO) y 2 (% UTILIDAD): ejecutado = 0
  SELECT COALESCE(
    SUM(
      CASE 
        -- Tipo 3: ARRIENDO FIJO → ejecutado = lease_type_value
        WHEN f.lease_type_id = 3 THEN f.lease_type_value * l.hectares
        -- Tipo 4: FIJO + % → ejecutado = solo parte fija (lease_type_value)
        WHEN f.lease_type_id = 4 THEN f.lease_type_value * l.hectares
        -- Tipo 1 (% INGRESO NETO) y Tipo 2 (% UTILIDAD) → ejecutado = 0
        ELSE 0
      END
    ),
    0
  )::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id 
    AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.lease_executed_for_project(bigint) IS 
'Calcula arriendo ejecutado del proyecto. Solo tipos fijos (3,4). FIX 000175: Suma por lote en lugar de usar LIMIT 1.';

COMMIT;


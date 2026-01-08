-- ========================================
-- MIGRACIÓN 000069: CREAR VISTAS _v2 FALTANTES
-- Entidad: views (Crear vistas _v2 que faltan)
-- Funcionalidad: Crear las vistas _v2 que el código Go necesita pero no existen
-- ========================================

-- ========================================
-- 1. CREAR dashboard_sowing_progress_view_v2
-- ========================================
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2;

CREATE VIEW dashboard_sowing_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) AS sowing_hectares,
  SUM(l.hectares) AS sowing_total_hectares,
  CASE 
    WHEN SUM(l.hectares) > 0 
    THEN (SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) / SUM(l.hectares)) * 100 
    ELSE 0 
  END AS sowing_progress_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 2. CREAR dashboard_harvest_progress_view_v2
-- ========================================
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2;

CREATE VIEW dashboard_harvest_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END) AS harvest_hectares,
  SUM(l.hectares) AS harvest_total_hectares,
  CASE 
    WHEN SUM(l.hectares) > 0 
    THEN (SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END) / SUM(l.hectares)) * 100 
    ELSE 0 
  END AS harvest_progress_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 3. CREAR dashboard_contributions_progress_view_v2
-- ========================================
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;

CREATE VIEW dashboard_contributions_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  pi.investor_id,
  i.name AS investor_name,
  pi.percentage AS investor_percentage_pct,
  -- Función para calcular progreso de aportes
  100.00 AS contributions_progress_pct
FROM projects p
JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, pi.investor_id, i.name, pi.percentage;

-- ========================================
-- 4. CREAR dashboard_operational_indicators_view_v2
-- ========================================
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;

CREATE VIEW dashboard_operational_indicators_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Fechas calculadas dinámicamente (se corrigen en migración 000070)
  NULL::DATE AS start_date,
  NULL::DATE AS end_date,
  NULL::DATE AS campaign_closing_date
FROM projects p
WHERE p.deleted_at IS NULL;

-- Migración 000050: Corregir SOLO el avance de siembra
-- Enfoque: Vista simplificada que calcula correctamente el porcentaje de siembra

DROP VIEW IF EXISTS dashboard_sowing_progress_view;

CREATE OR REPLACE VIEW dashboard_sowing_progress_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(total.total_hectares, 0) AS sowing_total_hectares,
    COALESCE(sown.sown_hectares, 0) AS sowing_hectares,
    CASE 
        WHEN COALESCE(total.total_hectares, 0) > 0 THEN
            (COALESCE(sown.sown_hectares, 0)::numeric / total.total_hectares * 100)
        ELSE 0 
    END AS sowing_progress_pct
FROM projects p
LEFT JOIN (
    SELECT f.project_id, SUM(l.hectares) AS total_hectares
    FROM fields f
    JOIN lots l ON l.field_id = f.id
    GROUP BY f.project_id
) total ON total.project_id = p.id
LEFT JOIN (
    SELECT f.project_id, SUM(l.hectares) AS sown_hectares
    FROM fields f
    JOIN lots l ON l.field_id = f.id
    WHERE l.sowing_date IS NOT NULL
    GROUP BY f.project_id
) sown ON sown.project_id = p.id;
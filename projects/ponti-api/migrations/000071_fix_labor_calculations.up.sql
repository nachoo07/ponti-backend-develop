-- ========================================
-- MIGRACIÓN 000071: CORREGIR CÁLCULOS DE LABORES
-- Entidad: labors (Corregir cálculos de IVA, costos en pesos y totales netos)
-- Funcionalidad: Crear vista fix_labors_list con cálculos correctos
-- ========================================

-- ========================================
-- 1. CREAR VISTA FIX_LABORS_LIST CON CÁLCULOS CORREGIDOS
-- ========================================
CREATE OR REPLACE VIEW fix_labors_list AS
SELECT
    w.id AS workorder_id,
    w.number AS workorder_number,
    w.date,
    p.id AS project_id,
    f.id AS field_id,
    p.name AS project_name,
    f.name AS field_name,
    c.name AS crop_name,
    lb.name AS labor_name,
    lc.name AS category_name,
    w.contractor,
    w.effective_area AS surface_ha,
    lb.price AS cost_ha,
    lb.contractor_name,
    inv.name AS investor_name,
    
    -- Obtener valor del dólar promedio (usando función get_default_fx_rate si no hay datos específicos)
    COALESCE(pdv.average_value, get_default_fx_rate()) AS usd_avg_value,
    
    -- Cálculos corregidos según especificaciones:
    
    -- Total neto: Costo labor * superficie (en USD)
    (lb.price * w.effective_area) AS net_total,
    
    -- Total IVA: Usar porcentaje configurable desde app_parameters
    (lb.price * w.effective_area * get_iva_percentage()) AS total_iva,
    
    -- Costo U$/Ha en pesos: Costo Ha * dólar promedio (mostrar en pesos)
    (lb.price * COALESCE(pdv.average_value, get_default_fx_rate())) AS usd_cost_ha,
    
    -- Total U Neto en pesos: Total costo en pesos * Has
    (lb.price * COALESCE(pdv.average_value, get_default_fx_rate()) * w.effective_area) AS usd_net_total,
    
    -- Campos de factura
    i.id AS invoice_id,
    i.number AS invoice_number,
    i.company AS invoice_company,
    i.date AS invoice_date,
    i.status AS invoice_status
    
FROM workorders w
INNER JOIN projects p ON w.project_id = p.id
INNER JOIN fields f ON w.field_id = f.id
INNER JOIN crops c ON w.crop_id = c.id
INNER JOIN labors lb ON w.labor_id = lb.id
INNER JOIN categories lc ON lb.category_id = lc.id
INNER JOIN investors inv ON w.investor_id = inv.id
LEFT JOIN invoices i ON i.work_order_id = w.id
LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id 
    AND pdv.deleted_at IS NULL
WHERE w.deleted_at IS NULL
    AND p.deleted_at IS NULL;

-- ========================================
-- 2. COMENTARIOS EXPLICATIVOS
-- ========================================
-- Total IVA: Ahora usa get_iva_percentage() desde app_parameters (configurable)
-- Costo U$/Ha: Se muestra en pesos (USD * tipo de cambio)
-- Total U Neto: Total en pesos multiplicado por hectáreas
-- Se usa get_default_fx_rate() como fallback si no hay datos de project_dollar_values

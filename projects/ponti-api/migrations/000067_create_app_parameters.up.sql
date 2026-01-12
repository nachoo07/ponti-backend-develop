-- ========================================
-- MIGRACIÓN 000067: CREAR PARÁMETROS DE APLICACIÓN
-- Entidad: app_parameters (Parámetros unificados del sistema)
-- Funcionalidad: Unificar units, calc_values y otros datos hardcodeados
-- ========================================

-- ========================================
-- 1. CREAR TABLA DE PARÁMETROS DE APLICACIÓN
-- ========================================
CREATE TABLE IF NOT EXISTS app_parameters (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- 'decimal', 'integer', 'string', 'boolean'
    category VARCHAR(50) NOT NULL,  -- 'units', 'calculations', 'business_rules'
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- 2. INSERTAR DATOS UNIFICADOS
-- ========================================
INSERT INTO app_parameters (key, value, type, category, description) VALUES
-- UNIDADES (antes unit)
('unit_liters', 'Lt', 'string', 'units', 'Unit of measurement: Liters'),
('unit_kilos', 'Kg', 'string', 'units', 'Unit of measurement: Kilograms'),
('unit_hectares', 'Ha', 'string', 'units', 'Unit of measurement: Hectares'),

-- CÁLCULOS (antes calc_values)
('iva_percentage', '0.105', 'decimal', 'calculations', 'VAT percentage for labors (10.5%)'),
('campaign_closure_days', '30', 'integer', 'calculations', 'Days for campaign closure after end date'),
('default_fx_rate', '1.0000', 'decimal', 'calculations', 'Default exchange rate USD/USD')
ON CONFLICT (key) DO NOTHING;

-- ========================================
-- 3. CREAR FUNCIONES DE PARÁMETROS
-- ========================================

-- Función genérica para obtener parámetros
CREATE OR REPLACE FUNCTION get_app_parameter(p_key VARCHAR)
RETURNS VARCHAR AS $$
BEGIN
  RETURN (SELECT value FROM app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para obtener parámetros decimales
CREATE OR REPLACE FUNCTION get_app_parameter_decimal(p_key VARCHAR)
RETURNS DECIMAL AS $$
BEGIN
  RETURN (SELECT value::DECIMAL FROM app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para obtener parámetros enteros
CREATE OR REPLACE FUNCTION get_app_parameter_integer(p_key VARCHAR)
RETURNS INTEGER AS $$
BEGIN
  RETURN (SELECT value::INTEGER FROM app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para obtener porcentaje de IVA
CREATE OR REPLACE FUNCTION get_iva_percentage()
RETURNS DECIMAL AS $$
BEGIN
  RETURN get_app_parameter_decimal('iva_percentage');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para obtener días de cierre de campaña
CREATE OR REPLACE FUNCTION get_campaign_closure_days()
RETURNS INTEGER AS $$
BEGIN
  RETURN get_app_parameter_integer('campaign_closure_days');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para obtener tasa de cambio por defecto
CREATE OR REPLACE FUNCTION get_default_fx_rate()
RETURNS DECIMAL AS $$
BEGIN
  RETURN get_app_parameter_decimal('default_fx_rate');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ========================================
-- 4. ACTUALIZAR VISTA FIX_LABORS_LIST PARA USAR VALORES DE CÁLCULO
-- ========================================
DROP VIEW IF EXISTS fix_labors_list;

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
    
    -- Obtener valor del dólar promedio (usando función configurable)
    COALESCE(pdv.average_value, get_default_fx_rate()) AS usd_avg_value,
    
    -- Cálculos usando valores de cálculo configurables:
    
    -- Total neto: Costo labor * superficie (en USD)
    (lb.price * w.effective_area) AS net_total,
    
    -- Total IVA: Usar porcentaje configurable
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
-- 5. ACTUALIZAR FUNCIÓN DE CÁLCULO DE FECHA DE CIERRE
-- ========================================
CREATE OR REPLACE FUNCTION calculate_campaign_closing_date(end_date DATE)
RETURNS DATE AS $$
BEGIN
  IF end_date IS NULL THEN
    RETURN NULL;
  END IF;
  
  -- Usar días configurable
  RETURN end_date + (get_campaign_closure_days() || ' days')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 6. RECREAR VISTA OPERATIONAL INDICATORS CON VALORES DE CÁLCULO
-- ========================================
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;

CREATE VIEW dashboard_operational_indicators_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  w_min.min_date AS start_date,
  w_max.max_date AS end_date,
  calculate_campaign_closing_date(w_max.max_date) AS campaign_closing_date
FROM projects p
LEFT JOIN (
  SELECT project_id, MIN(date) as min_date 
  FROM workorders 
  WHERE deleted_at IS NULL 
  GROUP BY project_id
) w_min ON w_min.project_id = p.id
LEFT JOIN (
  SELECT project_id, MAX(date) as max_date 
  FROM workorders 
  WHERE deleted_at IS NULL 
  GROUP BY project_id
) w_max ON w_max.project_id = p.id
WHERE p.deleted_at IS NULL;

-- ========================================
-- 7. CORREGIR VISTA CONTRIBUTIONS PROGRESS PARA CALCULAR SUMA REAL
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
  -- Calcular la suma real de porcentajes de participación (debe ser 100%)
  SUM(pi.percentage) OVER (PARTITION BY p.id) AS contributions_progress_pct
FROM projects p
JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

-- ========================================
-- 8. CREAR ÍNDICES PARA OPTIMIZACIÓN
-- ========================================
CREATE INDEX IF NOT EXISTS idx_app_parameters_key ON app_parameters(key);
CREATE INDEX IF NOT EXISTS idx_app_parameters_category ON app_parameters(category);

-- ========================================
-- 9. DOCUMENTACIÓN
-- ========================================
-- Esta migración unifica los siguientes datos:
-- 1. UNIDADES: unit_liters, unit_kilos, unit_hectares
-- 2. CÁLCULOS: iva_percentage, campaign_closure_days, default_fx_rate
-- 3. FUNCIONES: get_app_parameter, get_app_parameter_decimal, get_app_parameter_integer
--
-- Para cambiar estos valores, actualizar la tabla app_parameters:
-- UPDATE app_parameters SET value = '0.21' WHERE key = 'iva_percentage';
-- UPDATE app_parameters SET value = '45' WHERE key = 'campaign_closure_days';
-- UPDATE app_parameters SET value = '1.2000' WHERE key = 'default_fx_rate';

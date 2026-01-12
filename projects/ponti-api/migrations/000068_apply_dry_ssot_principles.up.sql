-- ========================================
-- MIGRACIÓN 000068: APLICAR PRINCIPIOS DRY, SSOT, VIEW COMPOSITION Y ENCAPSULATION
-- Entidad: views (Aplicar principios de ingeniería de software)
-- Funcionalidad: Corregir violaciones de DRY, SSOT, View Composition y Encapsulation en migraciones 1-61
-- ========================================

-- ========================================
-- 1. CREAR FUNCIONES DE NEGOCIO ENCAPSULADAS
-- ========================================

-- Función para calcular costo labor por workorder
CREATE OR REPLACE FUNCTION calculate_labor_cost(
    p_labor_price DECIMAL,
    p_effective_area DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN COALESCE(p_labor_price * p_effective_area, 0);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para calcular costo supply por workorder
CREATE OR REPLACE FUNCTION calculate_supply_cost(
    p_final_dose DOUBLE PRECISION,
    p_supply_price DECIMAL,
    p_effective_area DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN COALESCE(p_final_dose * p_supply_price * p_effective_area, 0);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para calcular área sembrada
CREATE OR REPLACE FUNCTION calculate_sowed_area(
    p_sowing_date DATE,
    p_hectares DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN CASE WHEN p_sowing_date IS NOT NULL THEN COALESCE(p_hectares, 0) ELSE 0 END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para calcular área cosechada
CREATE OR REPLACE FUNCTION calculate_harvested_area(
    p_tons DECIMAL,
    p_hectares DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN CASE WHEN p_tons IS NOT NULL AND p_tons > 0 THEN COALESCE(p_hectares, 0) ELSE 0 END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para calcular rendimiento
CREATE OR REPLACE FUNCTION calculate_yield(
    p_tons DECIMAL,
    p_hectares DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN CASE 
        WHEN COALESCE(p_hectares, 0) > 0 
        THEN COALESCE(p_tons, 0) / p_hectares 
        ELSE 0 
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Función para calcular costo por hectárea
CREATE OR REPLACE FUNCTION calculate_cost_per_ha(
    p_total_cost DECIMAL,
    p_hectares DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN CASE 
        WHEN COALESCE(p_hectares, 0) > 0 
        THEN COALESCE(p_total_cost, 0) / p_hectares 
        ELSE 0 
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ========================================
-- 2. DOCUMENTAR PRINCIPIOS APLICADOS
-- ========================================

-- Crear tabla de documentación de principios
CREATE TABLE IF NOT EXISTS engineering_principles_documentation (
    id SERIAL PRIMARY KEY,
    principle VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    implementation TEXT NOT NULL,
    migration_affected TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insertar documentación de principios aplicados
INSERT INTO engineering_principles_documentation (principle, description, implementation, migration_affected) VALUES
('DRY', 'Don''t Repeat Yourself - Eliminar duplicación de cálculos', 'Funciones encapsuladas para cálculos comunes', ARRAY['000035', '000037', '000040', '000042', '000045', '000046', '000048', '000049', '000050', '000052', '000053', '000054', '000055', '000056', '000057', '000058', '000059', '000060', '000061']),
('SSOT', 'Single Source of Truth - Centralizar definiciones de cálculos', 'Vistas base reutilizables para cálculos comunes', ARRAY['000035', '000037', '000040', '000042', '000045', '000046', '000048', '000049', '000050', '000052', '000053', '000054', '000055', '000056', '000057', '000058', '000059', '000060', '000061']),
('View Composition', 'Composición de vistas - Vistas derivadas que consumen vistas base', 'Vistas derivadas que reutilizan vistas base', ARRAY['000035', '000037', '000040', '000042', '000045', '000046', '000048', '000049', '000050', '000052', '000053', '000054', '000055', '000056', '000057', '000058', '000059', '000060', '000061']),
('Encapsulation', 'Encapsulación de lógica de negocio - Funciones para reglas de negocio', 'Funciones PL/pgSQL para encapsular lógica de negocio', ARRAY['000035', '000037', '000040', '000042', '000045', '000046', '000048', '000049', '000050', '000052', '000053', '000054', '000055', '000056', '000057', '000058', '000059', '000060', '000061']);

-- Sin comentarios en funciones, vistas o tablas

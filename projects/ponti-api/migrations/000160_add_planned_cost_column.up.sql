-- Agregar columna planned_cost solo si no existe (idempotente)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'projects' 
        AND column_name = 'planned_cost'
    ) THEN
        ALTER TABLE projects ADD COLUMN planned_cost NUMERIC(12,2);
    END IF;
END $$;
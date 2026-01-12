-- Eliminar columna planned_cost solo si existe (idempotente)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'projects' 
        AND column_name = 'planned_cost'
    ) THEN
        ALTER TABLE projects DROP COLUMN planned_cost;
    END IF;
END $$;
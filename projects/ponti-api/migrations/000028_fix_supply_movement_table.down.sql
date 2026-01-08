ALTER TABLE supply_movements
    ALTER COLUMN quantity TYPE DOUBLE PRECISION;

ALTER TABLE supply_movements
    DROP CONSTRAINT chk_movement_type;

ALTER TABLE supply_movements
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN created_by,
    DROP COLUMN updated_by,
    DROP COLUMN deleted_by;

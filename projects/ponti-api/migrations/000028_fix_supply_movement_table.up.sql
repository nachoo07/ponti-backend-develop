ALTER TABLE supply_movements
    ALTER COLUMN quantity TYPE NUMERIC(15,3) USING quantity::NUMERIC(15,3);

ALTER TABLE supply_movements
    ALTER COLUMN movement_type TYPE TEXT,
    ADD CONSTRAINT chk_movement_type CHECK (
        movement_type IN ('Stock', 'Movimiento interno', 'Remito oficial')
    );

ALTER TABLE supply_movements
    ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN created_by BIGINT,
    ADD COLUMN updated_by BIGINT,
    ADD COLUMN deleted_by BIGINT;

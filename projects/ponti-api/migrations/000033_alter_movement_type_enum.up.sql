ALTER TABLE IF EXISTS supply_movements DROP CONSTRAINT IF EXISTS chk_supply_movements_movement_type;

ALTER TABLE IF EXISTS supply_movements
    ADD CONSTRAINT chk_supply_movements_movement_type CHECK (movement_type = ANY (ARRAY['Stock'::text, 'Movimiento interno'::text, 'Remito oficial'::text, 'Movimiento interno entrada'::text]));

ALTER TABLE supply_movements
ADD CONSTRAINT chk_movement_type
CHECK (movement_type IN ('Stock', 'Movimiento interno', 'Remito oficial', 'Movimiento interno entrada'));

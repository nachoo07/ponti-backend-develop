ALTER TABLE stocks
    ALTER COLUMN initial_units TYPE NUMERIC(15,3),
    ALTER COLUMN initial_units SET NOT NULL,
    ALTER COLUMN real_stock_units TYPE NUMERIC(15,3) USING real_stock_units::NUMERIC(15,3);

ALTER TABLE stocks
    ADD COLUMN units_entered BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN units_consumed BIGINT NOT NULL DEFAULT 0;

ALTER TABLE stocks
    DROP COLUMN units_entered,
    DROP COLUMN units_consumed;

ALTER TABLE stocks
    ALTER COLUMN initial_units DROP NOT NULL,
    ALTER COLUMN initial_units TYPE float4,
    ALTER COLUMN real_stock_units TYPE float4;

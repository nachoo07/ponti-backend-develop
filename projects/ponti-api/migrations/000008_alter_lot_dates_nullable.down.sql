UPDATE lot_dates SET sowing_date = NOW() WHERE sowing_date IS NULL;

ALTER TABLE lot_dates
  ALTER COLUMN sowing_date SET NOT NULL;

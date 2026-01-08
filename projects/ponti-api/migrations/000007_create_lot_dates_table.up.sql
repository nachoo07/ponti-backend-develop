CREATE TABLE lot_dates (
  id SERIAL PRIMARY KEY,
  lot_id BIGINT NOT NULL,
  sowing_date DATE NOT NULL,
  harvest_date DATE,
  sequence SMALLINT NOT NULL CHECK (sequence BETWEEN 1 AND 3),
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,

  CONSTRAINT unique_lot_dates UNIQUE (lot_id, sequence),
  CONSTRAINT fk_lot_dates_lot FOREIGN KEY (lot_id) REFERENCES lots(id) ON DELETE CASCADE
);

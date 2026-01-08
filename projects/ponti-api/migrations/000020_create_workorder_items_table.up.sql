CREATE TABLE workorder_items (
  id             BIGSERIAL PRIMARY KEY,
  workorder_id   BIGINT     NOT NULL REFERENCES workorders(id) ON UPDATE CASCADE ON DELETE CASCADE,
  supply_id      BIGINT     NOT NULL REFERENCES supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  total_used     NUMERIC(18,6) NOT NULL,
  final_dose     NUMERIC(18,6) NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at     TIMESTAMPTZ
);

CREATE TABLE types (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(250) UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT
);


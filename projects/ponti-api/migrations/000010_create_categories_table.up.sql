CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(250) NOT NULL,
  type_id INTEGER NOT NULL REFERENCES types(id),
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT
);

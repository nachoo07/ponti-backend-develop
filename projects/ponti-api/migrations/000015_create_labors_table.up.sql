
CREATE TABLE labors (
  id BIGSERIAL PRIMARY KEY,
  project_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  category_id INTEGER NOT NULL, 
  price NUMERIC(12,2) NOT NULL,
  contractor_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,

  CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES categories(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

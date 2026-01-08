CREATE TABLE admin_cost_investors (
  project_id BIGINT NOT NULL,
  investor_id BIGINT NOT NULL,
  percentage INTEGER NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,

  PRIMARY KEY (project_id, investor_id),

  CONSTRAINT fk_admin_cost_investors_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_admin_cost_investors_investor FOREIGN KEY (investor_id) REFERENCES investors(id)
);

CREATE TABLE field_investors (
  field_id BIGINT NOT NULL,
  investor_id BIGINT NOT NULL,
  percentage INTEGER NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,

  PRIMARY KEY (field_id, investor_id),

  CONSTRAINT fk_field_investors_field FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE,
  CONSTRAINT fk_field_investors_investor FOREIGN KEY (investor_id) REFERENCES investors(id)
);
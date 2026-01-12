-- 001_init_schema.up.sql
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    refresh_tokens TEXT[] DEFAULT ARRAY[]::TEXT[],
    id_rol INT NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    created_by INT NOT NULL,
    updated_by INT NOT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

CREATE TABLE campaigns (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE customers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE managers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE investors (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE crops (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE lease_types (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255)
);

CREATE TABLE projects (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  customer_id BIGINT NOT NULL,
  campaign_id BIGINT NOT NULL,
  admin_cost BIGINT NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255),

  CONSTRAINT fk_projects_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
  CONSTRAINT fk_projects_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
);

CREATE INDEX idx_projects_customer_id ON projects(customer_id);
CREATE INDEX idx_projects_campaign_id ON projects(campaign_id);

CREATE TABLE fields (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  project_id BIGINT NOT NULL,
  lease_type_id BIGINT NOT NULL,
  lease_type_percent DOUBLE PRECISION,
  lease_type_value DOUBLE PRECISION,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255),

  CONSTRAINT fk_fields_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_lease_type FOREIGN KEY (lease_type_id) REFERENCES lease_types(id)
);

CREATE INDEX idx_fields_project_id ON fields(project_id);

CREATE TABLE lots (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  field_id BIGINT NOT NULL,
  hectares DOUBLE PRECISION NOT NULL,
  previous_crop_id BIGINT NOT NULL,
  current_crop_id BIGINT NOT NULL,
  season VARCHAR(20) NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255),

  CONSTRAINT fk_lots_field FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE,
  CONSTRAINT fk_lots_previous_crop FOREIGN KEY (previous_crop_id) REFERENCES crops(id),
  CONSTRAINT fk_lots_current_crop FOREIGN KEY (current_crop_id) REFERENCES crops(id)
);

CREATE INDEX idx_lots_field_id ON lots(field_id);
CREATE INDEX idx_lots_previous_crop_id ON lots(previous_crop_id);
CREATE INDEX idx_lots_current_crop_id ON lots(current_crop_id);

CREATE TABLE project_investors (
  project_id BIGINT NOT NULL,
  investor_id BIGINT NOT NULL,
  percentage INTEGER NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255),

  PRIMARY KEY (project_id, investor_id),

  CONSTRAINT fk_project_investors_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_project_investors_investor FOREIGN KEY (investor_id) REFERENCES investors(id)
);

CREATE TABLE project_managers (
  project_id BIGINT NOT NULL,
  manager_id BIGINT NOT NULL,

  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  deleted_by VARCHAR(255),

  PRIMARY KEY (project_id, manager_id),

  CONSTRAINT fk_project_managers_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_project_managers_manager FOREIGN KEY (manager_id) REFERENCES managers(id)
);

CREATE TABLE labor_types (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_by VARCHAR(255)
);
CREATE TABLE labor_categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type_id INTEGER NOT NULL REFERENCES labor_types(id),
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_by VARCHAR(255)
);




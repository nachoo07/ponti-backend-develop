CREATE TABLE project_dollar_values (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    year BIGINT NOT NULL,
    month VARCHAR(20) NOT NULL,
    start_value NUMERIC(12,2) NOT NULL,
    end_value NUMERIC(12,2) NOT NULL,
    average_value NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT,

    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    UNIQUE (project_id, year, month)
);

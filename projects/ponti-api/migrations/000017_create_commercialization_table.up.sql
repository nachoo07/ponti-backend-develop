CREATE TABLE crop_commercializations (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    crop_id BIGINT NOT NULL,
    board_price NUMERIC(12,2) NOT NULL,
    freight_cost NUMERIC(12,2) NOT NULL,
    commercial_cost DOUBLE PRECISION NOT NULL,
    net_price NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT,
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id),
    CONSTRAINT fk_crop FOREIGN KEY (crop_id) REFERENCES crops(id)
);

CREATE INDEX idx_crop_commercializations_project_id ON crop_commercializations(project_id);
CREATE INDEX idx_crop_commercializations_deleted_at ON crop_commercializations(deleted_at);

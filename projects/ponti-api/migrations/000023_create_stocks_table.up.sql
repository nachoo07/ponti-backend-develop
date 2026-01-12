-- +goose Up
-- SQL to create stocks table
CREATE TABLE stocks (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    field_id BIGINT NOT NULL,
    supply_id BIGINT NOT NULL,
    investor_id BIGINT NOT NULL,
    close_date DATE,
    real_stock_units float4 NOT NULL,
    initial_units float4,
    year_period INTEGER NOT NULL,
    month_period INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT,
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_field FOREIGN KEY (field_id) REFERENCES fields(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES investors(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

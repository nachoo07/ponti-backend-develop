CREATE TABLE supply_movements (
    id BIGSERIAL PRIMARY KEY,
    stock_id BIGINT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    movement_type movement_type NOT NULL,
    movement_date TIMESTAMP NOT NULL,
    reference_number TEXT NOT NULL,
    is_entry BOOLEAN NOT NULL,
    project_id BIGINT NOT NULL,
    field_id BIGINT NOT NULL,
    project_destination_id BIGINT NOT NULL,
    supply_id BIGINT NOT NULL,
    investor_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,


    CONSTRAINT fk_supply FOREIGN KEY (supply_id)
        REFERENCES supplies (id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    CONSTRAINT fk_investor FOREIGN KEY (investor_id)
        REFERENCES investors (id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    CONSTRAINT fk_provider FOREIGN KEY (provider_id)
        REFERENCES providers (id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);
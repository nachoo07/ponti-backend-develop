CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    work_order_id BIGINT NOT NULL UNIQUE,
    number VARCHAR NOT NULL,
    company VARCHAR(100) NOT NULL,
    date TIMESTAMP NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    created_by    BIGINT,
    updated_by    BIGINT,
    deleted_by    BIGINT,

    CONSTRAINT fk_invoices_work_order FOREIGN KEY (work_order_id) REFERENCES workorders(id) ON DELETE CASCADE
);

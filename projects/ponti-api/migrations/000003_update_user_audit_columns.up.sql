-- Alterar todas las tablas para cambiar las columnas de auditoría a BIGINT y agregar restricciones de clave foránea

ALTER TABLE campaigns
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_campaigns_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_campaigns_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_campaigns_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE customers
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_customers_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_customers_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_customers_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE managers
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_managers_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_managers_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE investors
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_investors_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_investors_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE crops
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_crops_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_crops_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_crops_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE lease_types
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_lease_types_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_lease_types_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_lease_types_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE projects
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_projects_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_projects_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_projects_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE fields
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_fields_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_fields_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_fields_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE lots
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_lots_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_lots_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_lots_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE project_investors
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_project_investors_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_project_investors_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_project_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE project_managers
  ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT,
  ALTER COLUMN updated_by TYPE BIGINT USING updated_by::BIGINT,
  ALTER COLUMN deleted_by TYPE BIGINT USING deleted_by::BIGINT,
  ADD CONSTRAINT fk_project_managers_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  ADD CONSTRAINT fk_project_managers_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  ADD CONSTRAINT fk_project_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id);

ALTER TABLE campaigns
  DROP CONSTRAINT fk_campaigns_created_by,
  DROP CONSTRAINT fk_campaigns_updated_by,
  DROP CONSTRAINT fk_campaigns_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);

ALTER TABLE customers
  DROP CONSTRAINT fk_customers_created_by,
  DROP CONSTRAINT fk_customers_updated_by,
  DROP CONSTRAINT fk_customers_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE managers
  DROP CONSTRAINT fk_managers_created_by,
  DROP CONSTRAINT fk_managers_updated_by,
  DROP CONSTRAINT fk_managers_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE investors
  DROP CONSTRAINT fk_investors_created_by,
  DROP CONSTRAINT fk_investors_updated_by,
  DROP CONSTRAINT fk_investors_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE crops
  DROP CONSTRAINT fk_crops_created_by,
  DROP CONSTRAINT fk_crops_updated_by,
  DROP CONSTRAINT fk_crops_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE lease_types
  DROP CONSTRAINT fk_lease_types_created_by,
  DROP CONSTRAINT fk_lease_types_updated_by,
  DROP CONSTRAINT fk_lease_types_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE projects
  DROP CONSTRAINT fk_projects_created_by,
  DROP CONSTRAINT fk_projects_updated_by,
  DROP CONSTRAINT fk_projects_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE fields
  DROP CONSTRAINT fk_fields_created_by,
  DROP CONSTRAINT fk_fields_updated_by,
  DROP CONSTRAINT fk_fields_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE lots
  DROP CONSTRAINT fk_lots_created_by,
  DROP CONSTRAINT fk_lots_updated_by,
  DROP CONSTRAINT fk_lots_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE project_investors
  DROP CONSTRAINT fk_project_investors_created_by,
  DROP CONSTRAINT fk_project_investors_updated_by,
  DROP CONSTRAINT fk_project_investors_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
ALTER TABLE project_managers
  DROP CONSTRAINT fk_project_managers_created_by,
  DROP CONSTRAINT fk_project_managers_updated_by,
  DROP CONSTRAINT fk_project_managers_deleted_by,
  ALTER COLUMN created_by TYPE VARCHAR(255),
  ALTER COLUMN updated_by TYPE VARCHAR(255),
  ALTER COLUMN deleted_by TYPE VARCHAR(255);
  
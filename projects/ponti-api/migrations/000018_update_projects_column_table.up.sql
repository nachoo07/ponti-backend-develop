ALTER TABLE projects
  ALTER COLUMN admin_cost TYPE NUMERIC(12,2);

UPDATE projects
  SET admin_cost = admin_cost / 100.0;

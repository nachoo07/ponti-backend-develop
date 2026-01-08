UPDATE projects
  SET admin_cost = ROUND(admin_cost * 100);

ALTER TABLE projects
  ALTER COLUMN admin_cost TYPE BIGINT
  USING admin_cost::BIGINT;

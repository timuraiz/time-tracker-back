-- Remove color column and its index from projects table
DROP INDEX IF EXISTS idx_projects_color;
ALTER TABLE projects DROP COLUMN IF EXISTS color;
-- Add color column to projects table
ALTER TABLE projects
ADD COLUMN color VARCHAR(7) DEFAULT '#3B82F6';

-- Add index for color if needed for filtering
CREATE INDEX IF NOT EXISTS idx_projects_color ON projects(color);
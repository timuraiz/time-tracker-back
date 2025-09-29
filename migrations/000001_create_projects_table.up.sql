-- Create projects table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create index for user_id
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

-- Create index for deleted_at (for soft deletes)
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
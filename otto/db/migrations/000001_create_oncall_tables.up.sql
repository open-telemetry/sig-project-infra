-- Create oncall_users table
CREATE TABLE IF NOT EXISTS oncall_users (
    id TEXT PRIMARY KEY,
    github_username TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create oncall_rotations table
CREATE TABLE IF NOT EXISTS oncall_rotations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    repository TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create oncall_assignments table
CREATE TABLE IF NOT EXISTS oncall_assignments (
    id TEXT PRIMARY KEY,
    rotation_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    is_current BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rotation_id) REFERENCES oncall_rotations(id),
    FOREIGN KEY (user_id) REFERENCES oncall_users(id)
);

-- Create oncall_escalations table
CREATE TABLE IF NOT EXISTS oncall_escalations (
    id TEXT PRIMARY KEY,
    assignment_id TEXT NOT NULL,
    issue_number INTEGER,
    pr_number INTEGER,
    repository TEXT NOT NULL,
    status TEXT NOT NULL,
    escalation_time TIMESTAMP,
    resolution_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES oncall_assignments(id)
);

-- Create indices for faster lookups
CREATE INDEX IF NOT EXISTS idx_oncall_users_github_username ON oncall_users(github_username);
CREATE INDEX IF NOT EXISTS idx_oncall_rotations_repository ON oncall_rotations(repository);
CREATE INDEX IF NOT EXISTS idx_oncall_assignments_rotation_id ON oncall_assignments(rotation_id);
CREATE INDEX IF NOT EXISTS idx_oncall_assignments_user_id ON oncall_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_oncall_assignments_is_current ON oncall_assignments(is_current);
CREATE INDEX IF NOT EXISTS idx_oncall_escalations_assignment_id ON oncall_escalations(assignment_id);
CREATE INDEX IF NOT EXISTS idx_oncall_escalations_repository ON oncall_escalations(repository);
CREATE INDEX IF NOT EXISTS idx_oncall_escalations_status ON oncall_escalations(status);
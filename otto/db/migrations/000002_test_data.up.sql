-- Insert test users
INSERT OR IGNORE INTO oncall_users (id, github_username, name, email, is_active)
VALUES 
    ('user-1', 'tester1', 'Test User 1', 'test1@example.com', TRUE),
    ('user-2', 'tester2', 'Test User 2', 'test2@example.com', TRUE),
    ('user-3', 'tester3', 'Test User 3', 'test3@example.com', TRUE);

-- Insert test rotations
INSERT OR IGNORE INTO oncall_rotations (id, name, description, repository, is_active)
VALUES 
    ('rotation-1', 'Main Rotation', 'Primary on-call rotation', 'open-telemetry/sig-project-infra', TRUE),
    ('rotation-2', 'Secondary Rotation', 'Backup on-call rotation', 'open-telemetry/sig-project-infra', TRUE);

-- Insert test assignments
INSERT OR IGNORE INTO oncall_assignments (id, rotation_id, user_id, start_time, end_time, is_current)
VALUES 
    ('assignment-1', 'rotation-1', 'user-1', DATETIME('now', '-7 days'), DATETIME('now', '-1 days'), FALSE),
    ('assignment-2', 'rotation-1', 'user-2', DATETIME('now'), DATETIME('now', '+7 days'), TRUE),
    ('assignment-3', 'rotation-2', 'user-3', DATETIME('now'), DATETIME('now', '+7 days'), TRUE);
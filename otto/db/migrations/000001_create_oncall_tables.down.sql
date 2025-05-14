-- Drop tables in reverse order to respect foreign key constraints
DROP TABLE IF EXISTS oncall_escalations;
DROP TABLE IF EXISTS oncall_assignments;
DROP TABLE IF EXISTS oncall_rotations;
DROP TABLE IF EXISTS oncall_users;
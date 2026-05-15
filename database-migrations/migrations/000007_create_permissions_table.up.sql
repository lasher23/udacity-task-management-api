CREATE TABLE permissions (
    id UUID NOT NULL PRIMARY KEY,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    UNIQUE (resource, action)
);

INSERT INTO permissions (id, resource, action) VALUES
('a0000000-0000-0000-0000-000000000001', 'task',    'read'),
('a0000000-0000-0000-0000-000000000002', 'task',    'write'),
('a0000000-0000-0000-0000-000000000003', 'task',    'delete'),
('a0000000-0000-0000-0000-000000000004', 'user',    'read'),
('a0000000-0000-0000-0000-000000000005', 'user',    'write'),
('a0000000-0000-0000-0000-000000000006', 'user',    'delete'),
('a0000000-0000-0000-0000-000000000007', 'profile', 'read');

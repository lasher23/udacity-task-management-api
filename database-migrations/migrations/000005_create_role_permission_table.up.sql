CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission VARCHAR(255) NOT NULL,
    PRIMARY KEY (role_id, permission),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

INSERT INTO role_permissions(role_id, permission) VALUES 
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'task:read'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'task:write'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'task:delete'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'user:read'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'user:write'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'user:delete'),
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'profile:read'),
('2b3c4d5e-6789-01ab-cdef-2345678901bc', 'task:read'),
('2b3c4d5e-6789-01ab-cdef-2345678901bc', 'task:write'),
('2b3c4d5e-6789-01ab-cdef-2345678901bc', 'profile:read');
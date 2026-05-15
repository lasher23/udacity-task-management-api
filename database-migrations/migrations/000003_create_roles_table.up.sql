CREATE TABLE roles (
    id UUID NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

INSERT INTO roles(id, name, description) VALUES 
('1c2d3e4f-5678-90ab-cdef-1234567890ab', 'admin', 'Administrator role with full permissions'),
('2b3c4d5e-6789-01ab-cdef-2345678901bc', 'user', 'Regular user role with limited permissions');
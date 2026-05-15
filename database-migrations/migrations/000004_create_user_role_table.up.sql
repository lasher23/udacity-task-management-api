CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

INSERT INTO user_roles(user_id, role_id) VALUES 
('bd006d41-aded-4040-9934-2ba4e909ef9a', '1c2d3e4f-5678-90ab-cdef-1234567890ab'),
('bd006d41-aded-4040-9934-2ba4e909ef9a', '2b3c4d5e-6789-01ab-cdef-2345678901bc');
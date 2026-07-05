-- Seed a default store
INSERT INTO stores (id, name, address, timezone, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'HQ Default Store',
    '123 Main Street, Manila, Philippines',
    'Asia/Manila',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Seed an admin user (username: admin, password: admin123)
-- Bcrypt hash of 'admin123' with cost 10 is '$2a$10$A6K7xlxU1749Gqjw7td.fOzPGu4sTD4Oy8ELEM3uD6wYRPrahrLIy'
INSERT INTO users (id, store_id, username, password_hash, role, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000000',
    'admin',
    '$2a$10$A6K7xlxU1749Gqjw7td.fOzPGu4sTD4Oy8ELEM3uD6wYRPrahrLIy',
    'admin',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (username) DO NOTHING;

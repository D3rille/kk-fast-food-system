-- Delete seeded admin user
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';

-- Delete seeded default store
DELETE FROM stores WHERE id = '00000000-0000-0000-0000-000000000000';

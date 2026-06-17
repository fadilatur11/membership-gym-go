INSERT INTO gyms (public_id, name, phone, email, address, timezone, currency, status)
VALUES ('00000000-0000-4000-8000-000000000001', 'Gym Sehat', '08123456789', 'admin@gym.com', 'Jl. Contoh No. 1', 'Asia/Jakarta', 'IDR', 'active')
ON CONFLICT (public_id) DO NOTHING;

-- Password hash is for "password"; regenerate with the app if your bcrypt version differs.
INSERT INTO users (public_id, gym_id, name, email, password_hash, role, is_active)
SELECT '00000000-0000-4000-8000-000000000002', g.id, 'Owner Gym', 'owner@gym.com', '$2y$10$8kkG1XhZ2cl9VD67g07PwuDKwg4gnK34Itll6Hc11AMoCWr40CUdm', 'owner', true
FROM gyms g
WHERE g.public_id = '00000000-0000-4000-8000-000000000001'
ON CONFLICT (gym_id, email) DO NOTHING;

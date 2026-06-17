CREATE TABLE saas_plans (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    code VARCHAR NOT NULL UNIQUE,
    name VARCHAR NOT NULL,
    description TEXT,
    duration_days INT NOT NULL,
    price BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR NOT NULL DEFAULT 'IDR',
    billing_cycle VARCHAR NOT NULL DEFAULT 'monthly',
    features JSONB NOT NULL DEFAULT '[]'::jsonb,
    limits JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_saas_plans_duration CHECK (duration_days > 0),
    CONSTRAINT chk_saas_plans_price CHECK (price >= 0),
    CONSTRAINT chk_saas_plans_billing_cycle CHECK (billing_cycle IN ('trial', 'monthly', 'yearly'))
);
CREATE INDEX idx_saas_plans_active ON saas_plans(is_active);
CREATE INDEX idx_saas_plans_code ON saas_plans(code);

CREATE TABLE gym_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    saas_plan_id BIGINT NOT NULL REFERENCES saas_plans(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'active',
    auto_renew BOOLEAN NOT NULL DEFAULT FALSE,
    source VARCHAR NOT NULL DEFAULT 'owner',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_gym_subscriptions_status CHECK (status IN ('trialing', 'active', 'pending', 'past_due', 'expired', 'cancelled')),
    CONSTRAINT chk_gym_subscriptions_dates CHECK (end_date >= start_date)
);
CREATE INDEX idx_gym_subscriptions_gym ON gym_subscriptions(gym_id);
CREATE INDEX idx_gym_subscriptions_plan ON gym_subscriptions(saas_plan_id);
CREATE INDEX idx_gym_subscriptions_status ON gym_subscriptions(gym_id, status);
CREATE INDEX idx_gym_subscriptions_end_date ON gym_subscriptions(gym_id, end_date);
CREATE UNIQUE INDEX uniq_current_gym_subscription
ON gym_subscriptions(gym_id)
WHERE status IN ('trialing', 'active', 'pending', 'past_due');

CREATE TABLE gym_subscription_payments (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    gym_subscription_id BIGINT NOT NULL REFERENCES gym_subscriptions(id),
    invoice_no VARCHAR NOT NULL,
    payment_method VARCHAR NOT NULL DEFAULT 'free',
    amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR NOT NULL DEFAULT 'IDR',
    status VARCHAR NOT NULL DEFAULT 'paid',
    paid_at TIMESTAMP,
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_gym_subscription_payments_invoice UNIQUE (gym_id, invoice_no),
    CONSTRAINT chk_gym_subscription_payments_amount CHECK (amount >= 0),
    CONSTRAINT chk_gym_subscription_payments_method CHECK (payment_method IN ('free', 'cash', 'transfer', 'qris')),
    CONSTRAINT chk_gym_subscription_payments_status CHECK (status IN ('pending', 'paid', 'failed', 'cancelled', 'refunded'))
);
CREATE INDEX idx_gym_subscription_payments_gym ON gym_subscription_payments(gym_id);
CREATE INDEX idx_gym_subscription_payments_subscription ON gym_subscription_payments(gym_subscription_id);
CREATE INDEX idx_gym_subscription_payments_status ON gym_subscription_payments(gym_id, status);

INSERT INTO saas_plans (public_id, code, name, description, duration_days, price, currency, billing_cycle, features, limits, is_active, created_at, updated_at)
VALUES
(
    '00000000-0000-4000-8000-000000000101',
    'basic',
    'Basic',
    'Trial 30 hari untuk operasional membership utama.',
    30,
    0,
    'IDR',
    'trial',
    '["membership.members","membership.packages","membership.subscriptions","membership.payments","member.qrcode","member.checkin"]'::jsonb,
    '{"max_members":100,"max_staff":2,"v2_workout":false,"reports":"none"}'::jsonb,
    true,
    NOW(),
    NOW()
),
(
    '00000000-0000-4000-8000-000000000102',
    'premium',
    'Premium',
    'Paket 30 hari untuk operasional gym lengkap tanpa fitur workout V2 lanjutan.',
    30,
    200000,
    'IDR',
    'monthly',
    '["membership.members","membership.packages","membership.subscriptions","membership.payments","member.qrcode","member.checkin","staff.users","expenses","reminders","reports.dashboard","reports.financial"]'::jsonb,
    '{"max_members":500,"max_staff":10,"v2_workout":false,"reports":"financial"}'::jsonb,
    true,
    NOW(),
    NOW()
),
(
    '00000000-0000-4000-8000-000000000103',
    'pro',
    'Pro',
    'Paket 30 hari dengan akses semua fitur termasuk workout V2 dan audit.',
    30,
    300000,
    'IDR',
    'monthly',
    '["all"]'::jsonb,
    '{"max_members":-1,"max_staff":-1,"v2_workout":true,"reports":"all"}'::jsonb,
    true,
    NOW(),
    NOW()
)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    duration_days = EXCLUDED.duration_days,
    price = EXCLUDED.price,
    currency = EXCLUDED.currency,
    billing_cycle = EXCLUDED.billing_cycle,
    features = EXCLUDED.features,
    limits = EXCLUDED.limits,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

INSERT INTO gym_subscriptions (public_id, gym_id, saas_plan_id, start_date, end_date, status, auto_renew, source, created_by, created_at, updated_at)
SELECT gen_random_uuid(), g.id, p.id, CURRENT_DATE, CURRENT_DATE + (p.duration_days - 1), 'trialing', false, 'system_seed', u.id, NOW(), NOW()
FROM gyms g
JOIN saas_plans p ON p.code = 'basic'
LEFT JOIN LATERAL (
    SELECT id FROM users WHERE users.gym_id = g.id AND users.role = 'owner' ORDER BY id ASC LIMIT 1
) u ON true
WHERE NOT EXISTS (
    SELECT 1 FROM gym_subscriptions gs
    WHERE gs.gym_id = g.id AND gs.status IN ('trialing', 'active', 'pending', 'past_due')
);

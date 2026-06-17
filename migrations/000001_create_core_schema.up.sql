CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE gyms (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    name VARCHAR NOT NULL,
    phone VARCHAR,
    email VARCHAR,
    address TEXT,
    timezone VARCHAR DEFAULT 'Asia/Jakarta',
    currency VARCHAR DEFAULT 'IDR',
    status VARCHAR NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    password_hash VARCHAR NOT NULL,
    role VARCHAR NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_users_gym_email UNIQUE (gym_id, email)
);
CREATE INDEX idx_users_gym_role ON users(gym_id, role);
CREATE INDEX idx_users_gym_active ON users(gym_id, is_active);

CREATE TABLE members (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_code VARCHAR NOT NULL,
    full_name VARCHAR NOT NULL,
    phone VARCHAR,
    email VARCHAR,
    gender VARCHAR,
    birth_date DATE,
    address TEXT,
    emergency_contact_name VARCHAR,
    emergency_contact_phone VARCHAR,
    joined_at DATE NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_members_gym_code UNIQUE (gym_id, member_code)
);
CREATE INDEX idx_members_gym_name ON members(gym_id, full_name);
CREATE INDEX idx_members_gym_status ON members(gym_id, status);
CREATE INDEX idx_members_gym_phone ON members(gym_id, phone);

CREATE TABLE membership_packages (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    duration_days INT NOT NULL,
    price BIGINT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_packages_gym_name ON membership_packages(gym_id, name);
CREATE INDEX idx_packages_gym_active ON membership_packages(gym_id, is_active);

CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    membership_package_id BIGINT NOT NULL REFERENCES membership_packages(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'active',
    source VARCHAR DEFAULT 'manual',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_subscriptions_gym_member ON subscriptions(gym_id, member_id);
CREATE INDEX idx_subscriptions_gym_status ON subscriptions(gym_id, status);
CREATE INDEX idx_subscriptions_gym_end_date ON subscriptions(gym_id, end_date);
CREATE INDEX idx_subscriptions_gym_member_status ON subscriptions(gym_id, member_id, status);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    subscription_id BIGINT REFERENCES subscriptions(id),
    invoice_no VARCHAR NOT NULL,
    payment_type VARCHAR NOT NULL DEFAULT 'membership',
    payment_method VARCHAR NOT NULL,
    package_price BIGINT NOT NULL DEFAULT 0,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    final_amount BIGINT NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'paid',
    paid_at TIMESTAMP NOT NULL,
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_payments_gym_invoice UNIQUE (gym_id, invoice_no)
);
CREATE INDEX idx_payments_gym_paid_at ON payments(gym_id, paid_at);
CREATE INDEX idx_payments_gym_status ON payments(gym_id, status);
CREATE INDEX idx_payments_gym_member ON payments(gym_id, member_id);
CREATE INDEX idx_payments_gym_method ON payments(gym_id, payment_method);

CREATE TABLE expense_categories (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_expense_categories_gym_name UNIQUE (gym_id, name)
);
CREATE INDEX idx_expense_categories_gym_active ON expense_categories(gym_id, is_active);

CREATE TABLE expenses (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    expense_category_id BIGINT NOT NULL REFERENCES expense_categories(id),
    title VARCHAR NOT NULL,
    description TEXT,
    amount BIGINT NOT NULL,
    expense_date DATE NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'approved',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expenses_gym_date ON expenses(gym_id, expense_date);
CREATE INDEX idx_expenses_gym_status ON expenses(gym_id, status);
CREATE INDEX idx_expenses_gym_category ON expenses(gym_id, expense_category_id);

CREATE TABLE member_qrcodes (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    qr_token VARCHAR UNIQUE NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'active',
    generated_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_qrcodes_gym_member ON member_qrcodes(gym_id, member_id);
CREATE INDEX idx_qrcodes_gym_status ON member_qrcodes(gym_id, status);
CREATE UNIQUE INDEX uniq_active_qrcode_per_member ON member_qrcodes(gym_id, member_id) WHERE status = 'active';

CREATE TABLE member_checkins (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    subscription_id BIGINT REFERENCES subscriptions(id),
    checkin_date DATE NOT NULL,
    checkin_at TIMESTAMP NOT NULL,
    source VARCHAR NOT NULL DEFAULT 'qr',
    status VARCHAR NOT NULL DEFAULT 'valid',
    scanned_by BIGINT REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_checkins_gym_date ON member_checkins(gym_id, checkin_date);
CREATE INDEX idx_checkins_gym_member ON member_checkins(gym_id, member_id);
CREATE INDEX idx_checkins_gym_member_date ON member_checkins(gym_id, member_id, checkin_date);
CREATE INDEX idx_checkins_gym_status ON member_checkins(gym_id, status);
CREATE INDEX idx_checkins_gym_source ON member_checkins(gym_id, source);
CREATE UNIQUE INDEX uniq_valid_checkin_per_member_per_day
ON member_checkins (gym_id, member_id, checkin_date)
WHERE status = 'valid';

CREATE TABLE reminder_rules (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    days_before_expiry INT NOT NULL,
    channel VARCHAR NOT NULL DEFAULT 'whatsapp',
    message_template TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reminder_rules_gym_active ON reminder_rules(gym_id, is_active);

CREATE TABLE reminder_logs (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id),
    reminder_rule_id BIGINT NOT NULL REFERENCES reminder_rules(id),
    channel VARCHAR NOT NULL DEFAULT 'whatsapp',
    recipient VARCHAR NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMP,
    provider_message_id VARCHAR,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reminder_logs_gym_status ON reminder_logs(gym_id, status);
CREATE INDEX idx_reminder_logs_gym_created ON reminder_logs(gym_id, created_at);
CREATE INDEX idx_reminder_logs_subscription_rule_recipient ON reminder_logs(subscription_id, reminder_rule_id, recipient);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    user_id BIGINT REFERENCES users(id),
    action VARCHAR NOT NULL,
    entity_type VARCHAR NOT NULL,
    entity_id BIGINT,
    payload JSON,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_logs_gym_created ON audit_logs(gym_id, created_at);
CREATE INDEX idx_audit_logs_gym_entity ON audit_logs(gym_id, entity_type);
CREATE INDEX idx_audit_logs_gym_user ON audit_logs(gym_id, user_id);

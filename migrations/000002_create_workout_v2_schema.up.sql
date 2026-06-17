CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    code VARCHAR NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_roles_gym_code UNIQUE (gym_id, code)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id BIGINT REFERENCES roles(id);
CREATE INDEX IF NOT EXISTS idx_users_gym_role_id ON users(gym_id, role_id);

INSERT INTO roles (public_id, gym_id, name, code, description, is_active, created_at, updated_at)
SELECT gen_random_uuid(), g.id, initcap(v.code), v.code, v.description, true, NOW(), NOW()
FROM gyms g
CROSS JOIN (VALUES
    ('owner', 'Full owner access'),
    ('admin', 'Operational admin access'),
    ('cashier', 'Cashier access'),
    ('trainer', 'Trainer access')
) AS v(code, description)
ON CONFLICT (gym_id, code) DO NOTHING;

UPDATE users u
SET role_id = r.id
FROM roles r
WHERE r.gym_id = u.gym_id
  AND r.code = u.role
  AND u.role_id IS NULL;

CREATE TABLE muscle_groups (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    code VARCHAR NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_muscle_groups_gym_code UNIQUE (gym_id, code)
);
CREATE INDEX idx_muscle_groups_gym_name ON muscle_groups(gym_id, name);
CREATE INDEX idx_muscle_groups_gym_active ON muscle_groups(gym_id, is_active);

CREATE TABLE workout_templates (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    name VARCHAR NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_workout_templates_gym_name ON workout_templates(gym_id, name);
CREATE INDEX idx_workout_templates_gym_active ON workout_templates(gym_id, is_active);

CREATE TABLE workout_template_days (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    workout_template_id BIGINT NOT NULL REFERENCES workout_templates(id),
    muscle_group_id BIGINT REFERENCES muscle_groups(id),
    day_of_week INT NOT NULL,
    title VARCHAR NOT NULL,
    description TEXT,
    is_rest_day BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_workout_template_days_day_of_week CHECK (day_of_week BETWEEN 1 AND 7),
    CONSTRAINT chk_workout_template_days_rest_muscle CHECK (is_rest_day = TRUE OR muscle_group_id IS NOT NULL)
);
CREATE INDEX idx_workout_template_days_gym_template ON workout_template_days(gym_id, workout_template_id);
CREATE INDEX idx_workout_template_days_gym_day ON workout_template_days(gym_id, day_of_week);
CREATE UNIQUE INDEX uniq_workout_template_day ON workout_template_days(gym_id, workout_template_id, day_of_week);

CREATE TABLE member_workout_schedules (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    workout_template_id BIGINT REFERENCES workout_templates(id),
    name VARCHAR NOT NULL DEFAULT 'Default Workout Schedule',
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_custom BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_member_workout_schedules_gym_member ON member_workout_schedules(gym_id, member_id);
CREATE INDEX idx_member_workout_schedules_gym_member_active ON member_workout_schedules(gym_id, member_id, is_active);
CREATE INDEX idx_member_workout_schedules_gym_active ON member_workout_schedules(gym_id, is_active);
CREATE UNIQUE INDEX uniq_active_member_workout_schedule ON member_workout_schedules(gym_id, member_id) WHERE is_active = true;

CREATE TABLE member_workout_schedule_days (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_workout_schedule_id BIGINT NOT NULL REFERENCES member_workout_schedules(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    muscle_group_id BIGINT REFERENCES muscle_groups(id),
    day_of_week INT NOT NULL,
    title VARCHAR NOT NULL,
    description TEXT,
    is_rest_day BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_member_workout_schedule_days_day_of_week CHECK (day_of_week BETWEEN 1 AND 7),
    CONSTRAINT chk_member_workout_schedule_days_rest_muscle CHECK (is_rest_day = TRUE OR muscle_group_id IS NOT NULL)
);
CREATE INDEX idx_member_workout_schedule_days_gym_member ON member_workout_schedule_days(gym_id, member_id);
CREATE INDEX idx_member_workout_schedule_days_gym_schedule ON member_workout_schedule_days(gym_id, member_workout_schedule_id);
CREATE INDEX idx_member_workout_schedule_days_gym_day ON member_workout_schedule_days(gym_id, day_of_week);
CREATE INDEX idx_member_workout_schedule_days_gym_member_day ON member_workout_schedule_days(gym_id, member_id, day_of_week);
CREATE UNIQUE INDEX uniq_member_workout_schedule_day ON member_workout_schedule_days(gym_id, member_workout_schedule_id, day_of_week);

CREATE TABLE member_workout_sessions (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    member_checkin_id BIGINT REFERENCES member_checkins(id),
    member_workout_schedule_day_id BIGINT REFERENCES member_workout_schedule_days(id),
    muscle_group_id BIGINT REFERENCES muscle_groups(id),
    workout_date DATE NOT NULL,
    title VARCHAR NOT NULL,
    notes TEXT,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    duration_seconds INT NOT NULL DEFAULT 0,
    status VARCHAR NOT NULL DEFAULT 'started',
    source VARCHAR NOT NULL DEFAULT 'member',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_member_workout_sessions_status CHECK (status IN ('started', 'paused', 'completed', 'cancelled', 'manual')),
    CONSTRAINT chk_member_workout_sessions_source CHECK (source IN ('member', 'admin', 'trainer'))
);
CREATE INDEX idx_member_workout_sessions_gym_member ON member_workout_sessions(gym_id, member_id);
CREATE INDEX idx_member_workout_sessions_gym_date ON member_workout_sessions(gym_id, workout_date);
CREATE INDEX idx_member_workout_sessions_gym_member_date ON member_workout_sessions(gym_id, member_id, workout_date);
CREATE INDEX idx_member_workout_sessions_gym_status ON member_workout_sessions(gym_id, status);
CREATE INDEX idx_member_workout_sessions_gym_muscle ON member_workout_sessions(gym_id, muscle_group_id);

CREATE TABLE member_weight_logs (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    gym_id BIGINT NOT NULL REFERENCES gyms(id),
    member_id BIGINT NOT NULL REFERENCES members(id),
    measured_date DATE NOT NULL,
    weight_kg NUMERIC(5,2) NOT NULL,
    notes TEXT,
    source VARCHAR NOT NULL DEFAULT 'member',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_member_weight_logs_weight_positive CHECK (weight_kg > 0),
    CONSTRAINT chk_member_weight_logs_source CHECK (source IN ('member', 'admin', 'trainer'))
);
CREATE INDEX idx_member_weight_logs_gym_member ON member_weight_logs(gym_id, member_id);
CREATE INDEX idx_member_weight_logs_gym_date ON member_weight_logs(gym_id, measured_date);
CREATE INDEX idx_member_weight_logs_gym_member_date ON member_weight_logs(gym_id, member_id, measured_date);

INSERT INTO muscle_groups (public_id, gym_id, code, name, description, is_active, created_at, updated_at)
SELECT gen_random_uuid(), g.id, v.code, v.name, v.description, true, NOW(), NOW()
FROM gyms g
CROSS JOIN (VALUES
    ('chest', 'Dada', 'Otot dada'),
    ('shoulder', 'Bahu', 'Otot bahu'),
    ('back', 'Punggung', 'Otot punggung'),
    ('legs', 'Kaki', 'Otot kaki'),
    ('arms', 'Lengan', 'Otot lengan'),
    ('biceps', 'Biceps', 'Otot biceps'),
    ('triceps', 'Triceps', 'Otot triceps'),
    ('abs', 'Perut', 'Otot perut'),
    ('cardio', 'Cardio', 'Latihan cardio'),
    ('full_body', 'Full Body', 'Latihan seluruh tubuh')
) AS v(code, name, description)
ON CONFLICT (gym_id, code) DO NOTHING;

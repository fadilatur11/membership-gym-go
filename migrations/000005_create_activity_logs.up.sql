CREATE TABLE activity_logs (
    id BIGSERIAL PRIMARY KEY,
    gym_id BIGINT REFERENCES gyms(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    request_payload JSONB,
    response JSONB,
    curl TEXT NOT NULL,
    status_code INT NOT NULL,
    status VARCHAR(10) NOT NULL,
    response_time BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_activity_logs_status_code CHECK (status_code BETWEEN 100 AND 599),
    CONSTRAINT chk_activity_logs_status CHECK (status IN ('success', 'failed')),
    CONSTRAINT chk_activity_logs_response_time CHECK (response_time >= 0)
);

CREATE INDEX idx_activity_logs_gym_created ON activity_logs(gym_id, created_at DESC);
CREATE INDEX idx_activity_logs_user_created ON activity_logs(user_id, created_at DESC);
CREATE INDEX idx_activity_logs_status_created ON activity_logs(status, created_at DESC);
CREATE INDEX idx_activity_logs_status_code ON activity_logs(status_code);

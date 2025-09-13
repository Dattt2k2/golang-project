CREATE TABLE expired_refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    refresh_token VARCHAR(256) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    UNIQUE
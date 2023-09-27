-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth (
    user_id BIGINT PRIMARY KEY,
    permissions INT NOT NULL,
    token VARCHAR(32) NOT NULL
);

-- TODO: maybe move proxy_ip to its own table
-- to reduce repeated data in future...
CREATE TABLE IF NOT EXISTS fingerprints (
    id SERIAL PRIMARY KEY,
    fingerprint VARCHAR(64) NOT NULL,
    proxy_ip VARCHAR(32) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE fingerprints;
DROP TABLE auth;
-- +goose StatementEnd

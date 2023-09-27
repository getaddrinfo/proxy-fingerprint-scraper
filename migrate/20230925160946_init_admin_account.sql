-- +goose Up
-- +goose StatementBegin
INSERT INTO auth (
    user_id,
    permissions,
    token
) VALUES (0, 15, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM auth WHERE user_id = 0;
-- +goose StatementEnd

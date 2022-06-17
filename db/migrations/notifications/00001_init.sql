-- +goose Up
-- +goose StatementBegin
CREATE TABLE notifications
(
    id       bigserial PRIMARY KEY,
    order_id bigint    NOT NULL,
    user_id  bigint    NOT NULL,
    ts       timestamp NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd

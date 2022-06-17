-- +goose Up
-- +goose StatementBegin
CREATE TABLE payments
(
    order_id bigint PRIMARY KEY,
    user_id  bigint          NOT NULL,
    total    decimal(32, 16) NOT NULL DEFAULT 0.0
);

CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'cancelled', 'failed');

CREATE TABLE payments_statuses
(
    order_id bigint PRIMARY KEY REFERENCES payments (order_id),
    status   payment_status NOT NULL DEFAULT 'pending'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments_statuses;
DROP TYPE IF EXISTS payment_status;
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd

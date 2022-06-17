-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders
(
    order_id      bigserial PRIMARY KEY,
    user_id       bigint          NOT NULL,
    total         decimal(32, 16) NOT NULL DEFAULT 0.0,
    delivery_date date            NOT NULL,
    email         varchar         NOT NULL
);

CREATE TABLE orders_items
(
    order_id   bigint NOT NULL REFERENCES orders (order_id),
    product_id bigint NOT NULL,
    quantity   bigint NOT NULL CHECK (quantity >= 0),

    PRIMARY KEY (order_id, product_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders_items;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd

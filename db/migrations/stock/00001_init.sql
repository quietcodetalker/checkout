-- +goose Up
-- +goose StatementBegin
CREATE TABLE quantities
(
    product_id bigint PRIMARY KEY,
    quantity   bigint NOT NULL CHECK (quantity >= 0)
);

CREATE TABLE reservations
(
    order_id   bigint NOT NULL,
    product_id bigint NOT NULL REFERENCES quantities (product_id),
    quantity   bigint NOT NULL CHECK (quantity >= 0),

    PRIMARY KEY (order_id, product_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reservations CASCADE;
DROP TABLE IF EXISTS quantities CASCADE;
-- +goose StatementEnd

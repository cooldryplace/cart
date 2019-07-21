-- +goose Up
CREATE TABLE carts (
  cart_id	SERIAL		PRIMARY KEY,
  user_id	INTEGER		NOT NULL,
  created_at 	TIMESTAMP 	NOT NULL,
  updated_at 	TIMESTAMP 	NOT NULL
);

CREATE TABLE line_items (
  item_id	SERIAL		PRIMARY KEY,
  cart_id	INTEGER 	REFERENCES carts,
  product_id	INTEGER		NOT NULL,
  quantity	INTEGER		NOT NULL,
  created_at 	TIMESTAMP 	NOT NULL,
  updated_at 	TIMESTAMP 	NOT NULL
);

-- +goose Down
DROP TABLE carts;
DROP TABLE line_items;

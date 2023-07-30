CREATE TABLE immutable_open_lot(
  open_lot_id serial primary key,
  lot_id uuid not null,
  cost_basis decimal not null, -- needed for asset split
  quantity decimal not null,
  trade_id int references trade(trade_id) not null unique,
  date date not null,
  created_at timestamp with time zone not null
);

ALTER TABLE closed_lot
ADD COLUMN open_lot_id int references immutable_open_lot(open_lot_id);
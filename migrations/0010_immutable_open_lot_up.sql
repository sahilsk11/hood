CREATE TABLE immutable_open_lot(
  open_lot_id serial primary key,
  cost_basis decimal not null, -- needed for asset split
  quantity decimal not null,
  trade_id int references trade(trade_id) not null,
  date timestamp with time zone not null,
  created_at timestamp with time zone not null,
  unique(trade_id, quantity) -- quantity should be the only field that changes (sells, asset splits)
);

alter TABLE cash
add column date timestamp with time zone not null;

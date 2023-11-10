CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE open_lot (
  open_lot_id serial primary key,
  cost_basis decimal not null, -- needed for asset split
  quantity decimal not null,
  trade_id int references trade(trade_id) not null unique,
  deleted_at timestamp with time zone, -- needed for asset split
  created_at timestamp with time zone not null,
  modified_at timestamp with time zone not null
);

CREATE TYPE gains_type as enum('SHORT_TERM', 'LONG_TERM');

CREATE TABLE closed_lot (
  closed_lot_id serial primary key,
  buy_trade_id int references trade(trade_id) not null,
  sell_trade_id int references trade(trade_id) not null,
  quantity decimal not null,
  realized_gains decimal not null,
  gains_type gains_type not null,
  created_at timestamp with time zone not null,
  modified_at timestamp with time zone not null,
  UNIQUE(buy_trade_id, sell_trade_id)
);

CREATE TABLE price (
  price_id serial primary key,
  symbol text not null,
  price decimal not null,
  updated_at timestamp with time zone not null
);

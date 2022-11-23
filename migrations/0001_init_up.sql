CREATE TYPE trade_action_type as enum('BUY', 'SELL');

CREATE TABLE trade (
  trade_id serial primary key,
  symbol text not null,
  action trade_action_type NOT NULL,
  quantity decimal not null,
  cost_basis decimal not null,
  date date not null,
  description text,
  created_at timestamp with time zone,
  modified_at timestamp with time zone,
  UNIQUE(symbol, action, quantity, cost_basis, date)
);

CREATE TABLE dividend (
  trade_id serial primary key,
  amount decimal not null,
  symbol text not null,
  date date not null,
  created_at timestamp with time zone,
  modified_at timestamp with time zone,
  UNIQUE(amount, symbol, date)
);

CREATE TYPE bank_activity_type as enum ('WITHDRAWAL', 'DEPOSIT');

CREATE TABLE bank_activity (
  activity_id serial primary key,
  amount decimal not null,
  activity_type bank_activity_type not null,
  date date not null,
  created_at timestamp with time zone,
  modified_at timestamp with time zone,
  UNIQUE(amount, activity_type, date)
);

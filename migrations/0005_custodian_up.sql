CREATE TYPE custodian_type as enum('ROBINHOOD', 'TDA');

ALTER TABLE trade add column custodian custodian_type not null default 'ROBINHOOD';

-- only purpose of this table is improved idempotency over table constraints --
CREATE TABLE tda_trade (
  tda_trade_id serial primary key,
  tda_transaction_id bigint not null unique,
  trade_id int references trade(trade_id)
);
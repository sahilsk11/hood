CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE bank_activity
rename to transfer;

ALTER TABLE open_lot
add column lot_id uuid not null default uuid_generate_v4()
add column date date not null default now(),
drop constraint open_lot_trade_id_key;

CREATE VIEW latest_open_lot AS
  SELECT a.* from open_lot a
  JOIN (
    select max(date), lot_id
    from open_lot
    group by lot_id
  ) b
  on a.date = b.date;

CREATE TABLE cash (
  cash_id serial primary key,
  amount decimal not null,
  custodian custodian_type not null,
  created_at timestamp with time zone not null
);
-- drop tables that are being moved to MDS
drop view latest_price;
drop table price;
drop table data_jockey_asset_metrics;

-- create tables for tracking user's and plaid items

-- represents a human
create table "user" (
  user_id uuid primary key,
  first_name text not null,
  middle_name text,
  last_name text not null,
  primary_email text not null,
  created_at timestamp with time zone not null
);

insert into "user" values (
  '398f8e70-2a98-446d-ab52-5ad414a3e9bc',
  'Sahil',
  null,
  'Kapur',
  'sahilkapur.a@gmail.com',
  now()
);

create table plaid_item(
  item_id uuid primary key,
  user_id uuid not null references "user"(user_id),
  plaid_item_id text not null,
  access_token text not null,
  created_at timestamp with time zone not null
);

CREATE TYPE account_type as enum('INDIVIDUAL', 'IRA', 'ROTH_IRA');

create table trading_account(
  trading_account_id uuid primary key,
  user_id uuid not null references "user"(user_id),
  custodian custodian_type not null,
  account_type account_type not null,
  created_at timestamp with time zone not null
);

insert into trading_account values (
  '2b0b4e9c-ef8c-424d-82ba-70e72c39dc19',
  '398f8e70-2a98-446d-ab52-5ad414a3e9bc',
  'SCHWAB',
  'INDIVIDUAL',
  now()
);

ALTER TABLE trade ADD COLUMN trading_account_id uuid not null references trading_account(trading_account_id) default '2b0b4e9c-ef8c-424d-82ba-70e72c39dc19';

ALTER TABLE trade ALTER COLUMN trading_account_id DROP DEFAULT;

-- note - this broke the robinhood trades, so i needed this too. not including in migrations

-- begin;
-- update trade
-- set trading_account_id = '08699bcd-7c8e-46c2-a161-bd1b30eaf4e8'
-- where custodian = 'ROBINHOOD';
-- rollback;

-- danger! can lose data here
alter table trade drop column custodian;
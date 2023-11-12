alter table plaid_account_metadata
rename to plaid_trading_account_metadata;

alter table plaid_trading_account_metadata
add column plaid_account_id text not null default '';

alter table plaid_trading_account_metadata
alter column plaid_account_id drop default;

create table plaid_trade_metadata (
  plaid_trade_metadata_id uuid primary key not null default uuid_generate_v4(),
  trade_id int references trade(trade_id),
  plaid_investment_transaction_id text not null
);

create type trade_source_type as enum ('MANUAL', 'PLAID', 'PLAID_INFERRED');

alter table trade add column source trade_source_type
not null default 'MANUAL';

alter table trade alter column source drop default;
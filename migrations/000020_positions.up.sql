alter table plaid_investment_holdings rename to position;

create type position_source_type as enum ('MANUAL', 'PLAID');

alter table
  position
add
  column source position_source_type not null default 'MANUAL';

alter table
  position
alter column
  source drop default;

-- add new column that represents how we
-- calculate holdings for an account
create type trading_account_data_source_type as enum ('TRADES', 'POSITIONS');

alter table
  trading_account
add
  column data_source trading_account_data_source_type not null default 'TRADES';

alter table
  trading_account
alter column
  data_source drop default;

-- gonna remove this constraint at a DB level
-- and attempt to enforce at code level

alter table trading_account
drop constraint trading_account_user_id_custodian_account_type_key;

alter table position
rename column plaid_investment_holdings_id to position_id;
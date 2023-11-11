alter table "user"
alter column user_id set default uuid_generate_v4();

alter table plaid_item
alter column item_id set default uuid_generate_v4();

alter table trading_account
alter column trading_account_id set default uuid_generate_v4();

alter type custodian_type
add value 'VANGUARD';
alter type custodian_type
add value 'WEALTHFRONT';
alter type custodian_type
add value 'BETTERMENT';
alter type custodian_type
add value 'MORGAN_STANLEY';
alter type custodian_type
add value 'E-TRADE';


alter type account_type
add value '401k';
alter type custodian_type
add value 'UNKNOWN';
alter type account_type
add value 'UNKNOWN';

create table plaid_account_metadata(
  plaid_account_metadata_id uuid primary key not null default uuid_generate_v4(),
  trading_account_id uuid not null references trading_account(trading_account_id),
  mask text,
  item_id uuid not null references plaid_item(item_id)
);

alter table trading_account
add column name text,
add unique(user_id, custodian, account_type);

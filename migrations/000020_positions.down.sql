alter table trading_account
add unique(user_id, custodian, account_type);


alter table trading_account drop column data_source;

drop type trading_account_data_source_type;

alter table position drop column source;

drop type positions_source_type;

alter table position
rename to plaid_investment_holdings;

alter table position
rename position_id to column plaid_investment_holdings_id;
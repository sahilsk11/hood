alter table position
add column deleted_at timestamp with time zone;

alter table position
drop constraint plaid_investment_holdings_ticker_trading_account_id_key;
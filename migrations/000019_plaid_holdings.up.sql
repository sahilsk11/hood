-- guess i didn't learn my lesson from immutable open lots
-- but i think we just update this table instead of rollup
create table plaid_investment_holdings (
  plaid_investments_holdings_id uuid primary key default uuid_generate_v4(),
  ticker text not null,
  trading_account_id uuid not null references trading_account(trading_account_id),
  total_cost_basis decimal not null,
  quantity decimal not null,
  created_at timestamp with time zone not null,
  unique(ticker, trading_account_id)
);
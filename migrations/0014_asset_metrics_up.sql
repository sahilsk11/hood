CREATE TABLE asset_metric(
  asset_metric_id serial primary key,
  symbol text not null,
  full_name text not null,
  price decimal not null,
  price_updated_at timestamp with time zone not null,
  earnings_per_share_annual_trailing decimal not null,
  earnings_per_share_annual_trailing_updated_at timestamp with time zone not null,
  dividend_yield_annual_trailing decimal not null,
  dividend_yield_annual_trailing_updated_at timestamp with time zone not null,
  pe_ratio_trailing decimal not null,
  book_value decimal not null,
  price_to_book_ratio decimal not null,
  shares_outstanding decimal not null,
  market_cap decimal not null
);
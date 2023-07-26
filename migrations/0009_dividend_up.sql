drop table dividend;

CREATE TABLE dividend(
  dividend_id serial primary key,
  symbol text not null,
  amount decimal not null,
  date date not null,
  custodian custodian_type not null,
  reinvestment_trade_id int references trade(trade_id),
  created_at timestamp with time zone
);
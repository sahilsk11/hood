CREATE VIEW vw_holding AS
  select symbol, sum(open_lot.quantity) AS "quantity", count(*) AS "num_lots"
  from open_lot
  inner join trade on trade.trade_id = open_lot.trade_id
  where deleted_at is null
  group by symbol;

CREATE VIEW vw_latest_price AS
  select a.*
  from price a
  join 
  ( select symbol, max(updated_at) as "updated_at"
  from price
  group by symbol) b
  on a.symbol = b.symbol and a.updated_at = b.updated_at;


CREATE VIEW vw_open_lot_position AS
  select
    open_lot.open_lot_id, trade.trade_id, trade.symbol, open_lot.quantity, trade.date AS "purchase_date", open_lot.cost_basis,
    (vw_latest_price.price - open_lot.cost_basis) * open_lot.quantity AS "unrealized_gains",
    CASE
      WHEN DATE_PART('day', now() - trade.date) >= 365 then 'LONG_TERM'
      else 'SHORT_TERM'
    end as "gains_type",
    vw_latest_price.price,
    vw_latest_price.updated_at as "price_updated_at"
  from open_lot
  inner join trade on trade.trade_id = open_lot.trade_id
  left join vw_latest_price on trade.symbol = vw_latest_price.symbol
  where deleted_at is null;

CREATE VIEW vw_unrealized_gain AS
  select symbol, sum(quantity) AS "quantity", sum(unrealized_gains) as "unrealized_gains", gains_type from vw_open_lot_position
  group by symbol, gains_type
  order by symbol;



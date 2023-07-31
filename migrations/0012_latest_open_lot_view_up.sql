drop view latest_open_lot;

CREATE VIEW current_open_lot AS WITH ranked_lots AS (
  SELECT 
    open_lot_id,
    cost_basis,
    quantity,
    trade_id,
    date,
    ROW_NUMBER() OVER (PARTITION BY trade_id ORDER BY date desc, open_lot_id desc) AS row_num
  FROM immutable_open_lot
)
SELECT 
	i.open_lot_id,
	i.cost_basis,
	i.quantity,
	i.trade_id,
	i.date
FROM ranked_lots i
WHERE row_num = 1 and i.quantity > 0;

/*
WITH ranked_lots AS (
  SELECT 
    open_lot_id,
    cost_basis,
    quantity,
    trade_id,
    date,
    ROW_NUMBER() OVER (PARTITION BY trade_id ORDER BY date desc, open_lot_id desc) AS row_num
  FROM immutable_open_lot
)
SELECT 
	t.symbol,
	sum(i.quantity),
	t.custodian
FROM ranked_lots i
inner join trade t on i.trade_id = t.trade_id
WHERE row_num = 1 and i.quantity > 0
group by t.symbol, t.custodian;
*/
drop view vw_unrealized_gain;
drop view vw_open_lot_position;
drop view vw_latest_price;


create view latest_price AS WITH ranked_prices AS (
  SELECT 
    price.*,
    ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY date desc, updated_at desc) AS row_num
  FROM price
)
SELECT 
	i.price_id, i.symbol, i.price, i.updated_at, i.date
FROM ranked_prices i
WHERE row_num = 1;
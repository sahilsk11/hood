ALTER TABLE transfer
rename to bank_activity;

ALTER TABLE open_lot
drop column lot_id,
drop column date,
add constraint open_lot_trade_id_key unique(trade_id);

drop view latest_open_lot;

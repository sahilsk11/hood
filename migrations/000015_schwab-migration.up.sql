ALTER TABLE tda_trade
DROP CONSTRAINT tda_trade_tda_transaction_id_key;

ALTER TABLE tda_trade
ALTER COLUMN tda_transaction_id DROP NOT NULL;
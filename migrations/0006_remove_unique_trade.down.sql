ALTER TABLE trade
add constraint trade_symbol_action_quantity_cost_basis_date_key
UNIQUE(symbol, action, quantity, cost_basis, date);
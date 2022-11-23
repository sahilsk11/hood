CREATE TABLE asset_split (
  asset_split_id serial primary key,
  symbol text not null,
  ratio int not null,
  date timestamp with time zone not null,
  created_at timestamp with time zone,
  UNIQUE(symbol, ratio, date)
);

CREATE TABLE applied_asset_split (
  applied_asset_split_id serial primary key,
  asset_split_id int references asset_split(asset_split_id) not null,
  open_lot_id int references open_lot(open_lot_id) not null,
  applied_at timestamp with time zone not null,
  UNIQUE(asset_split_id, open_lot_id)
);
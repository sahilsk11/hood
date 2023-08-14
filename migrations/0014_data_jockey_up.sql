-- i dont know which fields i need, so just store the whole json

CREATE TABLE data_jockey_asset_metrics(
  id serial primary key,
  json text not null,
  created_at timestamp with time zone not null
)
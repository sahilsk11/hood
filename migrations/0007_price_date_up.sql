ALTER TABLE price
ADD COLUMN date timestamp with time zone;

UPDATE price
set date = updated_at
where date is null;

alter table price
alter column date set not null;
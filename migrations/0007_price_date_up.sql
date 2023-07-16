ALTER TABLE price
ADD COLUMN date date;

UPDATE price
set date = updated_at::date
where date is null;

alter table price
alter column date set not null;

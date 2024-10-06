
create view if not exists vw_cache as
    select meter_id, consumption, datetime(last_entry, 'unixepoch', 'LOCALTIME') from cache;

insert into history values(
    '004-vw_cache.sql',
    'legers',
    'v0.0.0',
    'Added cache view to output date value.',
    date('2021-11-03')
);
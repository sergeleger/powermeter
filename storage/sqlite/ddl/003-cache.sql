
create table if not exists cache (
    meter_id integer primary key,
    consumption integer,
    last_entry integer
);

insert into history values(
    '003-cache.sql',
    'legers',
    'v0.0.0',
    'Added cache information to database.',
    date('2021-11-02')
);
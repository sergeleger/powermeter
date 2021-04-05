-- power contains the power measurements
create table if not exists power (
    meter_id integer,
    year integer,
    month integer,
    day integer,
    seconds integer,
    consumption integer
);

insert into history values(
    '001-power.sql',
    'legers',
    'v0.0.0',
    'Initial database creation',
    date('2021-03-14')
);

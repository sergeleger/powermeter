-- Index for faster filtering
CREATE INDEX if not exists power_meter_id_ymd on
    power(meter_id, year, month, day);

insert into history values(
    '002-indexes.sql',
    'legers',
    'v0.0.0',
    'Created indexes for power',
    date('2021-04-02')
);


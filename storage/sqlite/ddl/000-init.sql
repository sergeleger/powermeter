
-- history contains historical DDL updates performed to the databse.
create table if not exists history (
    file text,
    user text,
    version text,
    description text,
    date text
);

insert into history values(
    '000-init.sql',
    'legers',
    'v0.0.0',
    'Initial database creation',
    date('2021-03-14')
);

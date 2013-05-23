create table site (name varchar(10) primary key);

insert into site values ('coupang');
insert into site values ('tmon');
insert into site values ('wmp');
insert into site values ('groupon');

create table deal_daily_snapshot (
    site varchar(10),
    deal_id bigint,
    day date,
    created datetime not null,
    updated datetime not null,
    expired bool not null,
    adult bool not null,
    original_price int,
    discount_price int,
    num_sold int,
    description varchar(500),
    category varchar(100),
    subcategory varchar(100),
    locale varchar(200),
    primary key (site, deal_id, day));

create table option_daily_snapshot (
    site varchar(10),
    deal_id bigint,
    option_id bigint,
    day date,
    created datetime not null,
    updated datetime not null,
    price int,
    num_available int,
    num_sold int,
    description varchar(500),
    primary key (site, deal_id, option_id, day));

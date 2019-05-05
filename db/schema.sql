-- Database Schema

create table versions (
    applied timestamp without time zone default (now()  at time zone 'utc')
    , file varchar(150)
);

create table maps (
    map_id serial primary key 
    , region_name varchar(50)
    , created_at timestamp without time zone default (now() at time zone 'utc')
    , updated_at timestamp without time zone default (now() at time zone 'utc')
    , data json
);

create table corporations (
    corporation_id serial primary key
    , map_id integer references maps on delete cascade 
    , data json
    , name varchar(50)
);

create table cities (
    city_id serial primary key 
    , map_id integer references maps default NULL on delete cascade
    , city_name varchar(50) 
    , updated_at timestamp  without time zone default (now() at time zone 'utc')
    , data json 
    , corporation_id integer references corporations on delete set NULL default NULL
);

create table neighbouring_cities (
    neighbouring_cities serial primary key
    , from_city_id integer references cities(city_id) on delete cascade
    , to_city_id integer references cities(city_id) on delete cascade
);

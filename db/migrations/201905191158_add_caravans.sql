create table caravans (
    caravan_id serial primary key
    , origin_corporation_id integer references corporations on delete set NULL default NULL
    , target_corporation_id integer references corporations on delete set NULL default NULL
    , origin_city_id integer references cities(city_id) on delete cascade
    , target_city_id integer references cities(city_id) on delete cascade
    , state integer default 0
    , map_id integer references maps on delete cascade 
    , updated_at timestamp  without time zone default (now() at time zone 'utc')
    , data json
);

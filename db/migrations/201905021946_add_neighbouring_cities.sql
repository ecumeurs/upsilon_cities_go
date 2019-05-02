
create table neighbouring_cities (
    neighbouring_cities serial primary key
    , from_city_id integer references cities(city_id)
    , to_city_id integer references cities(city_id)
);
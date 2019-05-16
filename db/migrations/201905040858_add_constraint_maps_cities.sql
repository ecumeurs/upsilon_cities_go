
alter table cities
    drop constraint cities_map_id_fkey;

alter table cities
    add constraint cities_map_id_fkey
    foreign key (map_id)
    references maps (map_id)
    on delete cascade
    default NULL;

    
alter table neighbouring_cities
    drop constraint neighbouring_cities_to_city_id_fkey;

alter table neighbouring_cities
    add constraint neighbouring_cities_to_city_id_fkey
    foreign key (to_city_id)
    references cities (city_id)
    on delete cascade;

alter table neighbouring_cities
    drop constraint neighbouring_cities_from_city_id_fkey;

alter table neighbouring_cities
    add constraint neighbouring_cities_from_city_id_fkey
    foreign key (from_city_id)
    references cities (city_id)
    on delete cascade;

    
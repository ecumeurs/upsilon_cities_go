
create table corporations (
    corporation_id serial primary key
    , map_id integer references maps on delete cascade 
    , data json
    , name varchar(50)
);

alter table cities add column 
    corporation_id integer references corporations on delete set NULL default NULL;


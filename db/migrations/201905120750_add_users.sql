create table users (
    user_id serial primary key 
    , login varchar(50) unique
    , email varchar(50) unique
    , password varchar(256)
    , enabled boolean 
    , admin boolean
    , last_login timestamp without time zone default (now() at time zone 'utc')
    , data json -- dont know maybe will have user preferences and stuff like that ;)
);

create table user_logs (
    user_log_id  serial primary key
    , user_id integer references users(user_id) on delete cascade
    , message varchar(200) 
    , gravity integer 
    , inserted timestamp  without time zone default (now() at time zone 'utc')
    , acknowledged timestamp  without time zone default NULL
);



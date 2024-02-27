create table if not exists request (
    id serial primary key,
    method text not null,
    body text not null
);

create table if not exists response (
    id serial primary key,
    status_code int not null,
    body text not null
);

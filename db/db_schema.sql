create table if not exists request
(
    id           serial primary key,
    method       text  not null,
    path         text  not null,
    query_params jsonb not null,
    headers      jsonb not null,
    cookies      jsonb not null,
    content_type text  not null,
    body         text  not null
);

create table if not exists response
(
    id           serial primary key,
    request_id   int   not null,
    status_code  int   not null,
    headers      jsonb not null,
    content_type text  not null,
    body         text  not null,

    constraint fk_request_id foreign key (request_id) references request on delete cascade
);

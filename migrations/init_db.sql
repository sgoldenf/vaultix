create table if not exists users (
    id bigint primary key,
    hashed_password char(60) not null
);

create table if not exists passwords (
    id serial primary key,
    user_id bigint not null,
    service varchar(255) not null,
    login varchar(255) not null,
    encrypted_password bytea not null,
    constraint fk_user foreign key(user_id) references users(id),
    constraint passwords_user_service_login_uc unique (user_id, service, login)
);

create index if not exists user_id_idx on passwords (user_id);

create extension pg_trgm;
create schema mine;

create table mine.help (
  id serial primary key
);

create table person (
  id serial primary key,
  created timestamp default now(),
  updated timestamp default now()
);

create table name (
  id integer references person(id) not null,
  name text unique not null,
  updated timestamp default now(),
  created timestamp default now()
);

create index idx_okay on name(name, created);

create table age (
  id integer references person(id) not null ,
  age integer not null,
  updated timestamp default now(),
  created timestamp default now()
);

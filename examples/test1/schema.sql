create table person (
  id serial primary key,
  created timestamp default now()
);

create table name (
  id integer references person(id) not null,
  name text not null
);

create table age (
  id integer references person(id) not null ,
  age integer not null,
  created timestamp default now()
);

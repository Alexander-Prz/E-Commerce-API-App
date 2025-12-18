create table public.customer_games (
  id bigserial not null,
  customerid bigint not null,
  gameid bigint not null,
  purchased_at timestamp with time zone null default now(),
  constraint customer_games_pkey primary key (id),
  constraint customer_games_customerid_gameid_key unique (customerid, gameid),
  constraint customer_games_customerid_fkey foreign KEY (customerid) references customers (customerid),
  constraint customer_games_gameid_fkey foreign KEY (gameid) references games (gameid)
) TABLESPACE pg_default;

create table public.customers (
  customerid serial not null,
  authid integer not null,
  username character varying(20) null,
  fullname character varying(100) null,
  email character varying(150) not null,
  address text null,
  phone character varying(20) null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint customers_pkey primary key (customerid),
  constraint customers_authid_key unique (authid),
  constraint customers_email_key unique (email),
  constraint customers_authid_fkey foreign KEY (authid) references userauth (authid)
) TABLESPACE pg_default;

create table public.developers (
  developerid serial not null,
  developername character varying(150) not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  authid integer null,
  constraint developers_pkey primary key (developerid),
  constraint developers_authid_key unique (authid),
  constraint fk_developers_auth foreign KEY (authid) references userauth (authid)
) TABLESPACE pg_default;

create table public.email_verifications (
  authid integer not null,
  token text not null,
  expires_at timestamp without time zone not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  constraint email_verifications_token_key unique (token),
  constraint email_verifications_authid_fkey foreign KEY (authid) references userauth (authid)
) TABLESPACE pg_default;

create table public.gamegenres (
  gameid integer not null,
  genreid integer not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint gamegenres_pkey primary key (gameid, genreid),
  constraint gamegenres_gameid_fkey foreign KEY (gameid) references games (gameid),
  constraint gamegenres_genreid_fkey foreign KEY (genreid) references genres (genreid)
) TABLESPACE pg_default;

create table public.games (
  gameid serial not null,
  developerid integer not null,
  title character varying(200) not null,
  price numeric(10, 2) not null,
  releasedate date null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint games_pkey primary key (gameid),
  constraint games_developerid_fkey foreign KEY (developerid) references developers (developerid)
) TABLESPACE pg_default;

create table public.genres (
  genreid serial not null,
  genrename character varying(100) not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint genres_pkey primary key (genreid),
  constraint genres_genrename_key unique (genrename)
) TABLESPACE pg_default;

create table public.orderitems (
  orderitemid serial not null,
  orderid integer not null,
  gameid integer not null,
  quantity integer not null default 1,
  priceatpurchase numeric(10, 2) not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint orderitems_pkey primary key (orderitemid),
  constraint orderitems_gameid_fkey foreign KEY (gameid) references games (gameid),
  constraint orderitems_orderid_fkey foreign KEY (orderid) references orders (orderid)
) TABLESPACE pg_default;

create table public.orders (
  orderid serial not null,
  customerid integer not null,
  orderdate timestamp without time zone null default CURRENT_TIMESTAMP,
  totalprice numeric(10, 2) null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  constraint orders_pkey primary key (orderid),
  constraint orders_customerid_fkey foreign KEY (customerid) references customers (customerid)
) TABLESPACE pg_default;

create table public.paymentlogs (
  logid serial not null,
  paymentid integer not null,
  oldstatus character varying(20) null,
  newstatus character varying(20) null,
  changedat timestamp without time zone null default CURRENT_TIMESTAMP,
  constraint paymentlogs_pkey primary key (logid),
  constraint paymentlogs_paymentid_fkey foreign KEY (paymentid) references payments (paymentid)
) TABLESPACE pg_default;

create table public.paymentmethods (
  paymentmethodid serial not null,
  name character varying(100) not null,
  constraint paymentmethods_pkey primary key (paymentmethodid)
) TABLESPACE pg_default;

create table public.payments (
  paymentid serial not null,
  orderid integer not null,
  paymentmethodid integer not null,
  amountpaid numeric(10, 2) not null,
  paymentstatus character varying(20) not null default 'Pending'::character varying,
  createdat timestamp without time zone null default CURRENT_TIMESTAMP,
  paidat timestamp without time zone null,
  constraint payments_pkey primary key (paymentid),
  constraint payments_orderid_fkey foreign KEY (orderid) references orders (orderid),
  constraint payments_paymentmethodid_fkey foreign KEY (paymentmethodid) references paymentmethods (paymentmethodid)
) TABLESPACE pg_default;

create table public.userauth (
  authid serial not null,
  email character varying(150) not null,
  passwordhash text not null,
  role character varying(20) not null,
  created_at timestamp without time zone null default CURRENT_TIMESTAMP,
  deleted_at timestamp without time zone null,
  email_verified boolean not null default false,
  constraint userauth_pkey primary key (authid),
  constraint userauth_email_key unique (email),
  constraint userauth_role_check check (
    (
      (role)::text = any (
        (
          array[
            'admin'::character varying,
            'user'::character varying,
            'developer'::character varying
          ]
        )::text[]
      )
    )
  )
) TABLESPACE pg_default;
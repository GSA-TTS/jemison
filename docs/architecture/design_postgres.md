# postgres design

## the issue with the current design

The SQLite-based approach will have scaling issues in cloud.gov.

Specifically, cloud.gov puts a hard 7GB limit on a single compute instance. This means multiple (bad) things for an approach based on SQLite databases.

1. As we pack a DB, it must be smaller than 7GB.
2. If we want to `VACUUM` a DB, it must be smaller than ~2GB. (`VACUUM` may require 2x the size of the DB for cleanup and compression. This suggests a 2GB DB might require 4GB additional while undergoing a `VACUUM` before deployment.)
3. Large sites are clearly larger than 7GB (or 2GB)

So, what does this look like if we use Postgres? A few design ideas...

## overall: a data pipeline (again)

The FAC database design recently turned into a data pipeline. It's a smaller set of data (which makes copying fast/inexpensive), but we basically engage in a series of safe transformations to the data that let us do interesting things.

The rest of Jemison is designed this way, so why not the DB?

The first thing is to get the content that `fetch`/`walk`/`extract` develop into a database. From there, we can run jobs that do interesting things. The important thing is that content lands somewhere.

## three databases (possibly four)

The `queues` database currently is where `river` does all of its work. This is a high-traffic zone, and probably a bad place to put any other tables. It can be a small instance, but it probably should be all by itself. Let the queues thrash.

Then, we need to do a few things:

1. Track what URLs have been visited, and when
2. Get the content in, and ready for processing/indexing

Arguably... we could skip a table, and go straight to "what we need." But, that would imply writing custom code to go from S3->Pg. Better to have a generic "slurp the content into Pg," and then use SQL for the rest of the data pipeline. 

### constants / lookup tables

There are a few lookup tables that we want, to keep things small in the otherwise large tables. 

We could use `ENUM`. There are tradeoffs:

https://neon.tech/postgresql/postgresql-tutorial/postgresql-enum

Long-and-short, I think I'll stick to lookup tables. Extensible, simple, easy to maintain via migration.

> [!IMPORTANT]  
> The idea of using [domain64](domain64.md) notation for the content tables was developed after some of this content was written. I've come back around and updated the section regarding lookup tables to reflect this idea.

#### `schemes`

Huh. [Don't do this](https://wiki.postgresql.org/wiki/Don%27t_Do_This#Don.27t_use_serial).

```sql
create table schemes (
  id integer generated by default as identity primary key,
  scheme text,
  unique(scheme)
);

insert into schemes 
	(id, scheme) 
	values 
	(1, 'https'), (2, 'http'), (3, 'ftp')
	on conflict do nothing
;
```

#### `tlds`

A TLD should map to its `domain64` representation. 

```sql
create table tlds (
  id,
  tld text,
  unique(id),
  unique(tld),
  constraint domain64_tld check (id > 0 and id < 256)
);

insert into tlds
  (id, tld)
  values
  (1, 'gov'), (2, 'com'), (3, 'org'), (4, 'net')
  on conflict do nothing
;
```

These can be maintained as [jsonnet](https://jsonnet.org) files in the repository, and loaded on every deploy.

#### `content_type`

```sql
create table content_types (
  id integer generated by default as identity primary key,
  content_type text
  unique(content_type)
);

insert into content_types
  (id, content_type)
  values
  (1, 'application/octet-stream'), (2, 'text/html'), (3, 'application/pdf')
  on conflict do nothing
;
```

We will use [simplified/clean MIME types](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types/Common_types). Some get more complex when the webserver talks to us. 

#### `hosts`

I want to know the hosts I should know about. This is important for crawling: we never want to wander outside the space of what we are supposed to know. However, this should be information that is *derivative* of the domain64 information we're tracking.

From a domain64 perspective, these are the domains.

```sql
create table domains (
  id bigint,
  domain text,
  unique(id),
  unique(domain),
  constraint domain64_domain check (domain > 0 and domain <= select(x'FFFFFF'))
);

insert into hosts
  (id, host)
  values
  (...)
  on conflict do nothing
;
```

Again, this will be a Jsonnet file. The uniqueness constraint will help prevent re-ordering. New domains should only ever get new indexes, and we should never renumber. That is, if we remove a domain, we should *comment it out* in the Jsonnet, and "retire" the number. *Monotonically increasing* is a term that comes to mind.

#### `tags`

```sql
create table tags (
  id integer generated by default as identity primary key,
  tag text not null,
  unique(tag)
;

insert into tags
  (id, tag)
  values
  (1, 'title'),
  (2, 'p'),
  (3, 'div'),
  (...)
;
```

### the guestbook

With those tables in place...

```sql
create table if not exists guestbook (
  id bigserial primary key,
  last_modified timestamp not null, -- should this be nullable?
  last_fetched timestamp not null,
  next_fetch timestamp not null,
  scheme integer not null references schemes(id) default 1,
  host integer references hosts(id) not null,
  content_type integer references content_types(id) default 1, -- nullable?
  content_length integer not null default 0,
  path text not null,
  unique (host, path)
);
```

This used to have a `sha1` of the content, and the content length. The SHA1 is probably less useful than the ETag... and that might be debateable. Defaulting the `content_length` to 0 feels right... I'd rather always find a number than have `null`. 

#### a note about table ordering...

*re mi do do so*

https://docs.gitlab.com/ee/development/database/ordering_table_columns.html

The `guestbook` table (and tables that follow) will get big enough that padding will matter. Therefore, column ordering (8 word values, then 4, then variable) will matter a great deal for page alignment.  

#### `guestbook` is a lookup table

In many ways. But, the `host/path` combination now has a unique id. This means that we can use that `id` in other tables, and know that we're referring to a particular path on a particular host. For the content tables, we'll refer to the guestbook `id`. 

*What happens if we loose the guestbook table?*

Then we need to start a fresh crawl. That, or we back it up every now and then. We'll see what we do.

## the content

`extract` puts extracted content into S3, and we'll then pack that content into a table. The `raw_content` table is the first step of our SQL-based data pipeline. It will be in the same database as the `guestbook`, unless we ultimately decide that the cost of having it there is prohibitive (in performance or space).

```sql
CREATE TABLE raw_content (
  id BIGSERIAL PRIMARY KEY,
  host_path BIGINT references guestbook(id),
  tag TEXT default ,
  content TEXT 
)
```

Hm. Now, the `title` becomes a `tag`. A header becomes a `tag`. And, all the body content is by-tag. We can still prioritize/weight things by `tag`. (A `path` tag might even be a way to put all content, including paths, into one table.)

## the pipeline

From here, the question is "what kinds of search do we want to support?"

The `raw_content` table will end up being... at least 25M rows, possibly pushing closer to 100M rows. However, the idea here is not to work with this table directly. It is where we have an up-to-date version of the content of websites and PDFs (and other documents) in a location that is ready and amenable to batch, pipeline processing in SQL.

### one idea: inheritence.

https://www.postgresql.org/docs/current/tutorial-inheritance.html

We could define a searchable table as `gov`. 

```sql
create table gov (
  id ...,
  host_path ...,
  tag ...,
  content ...
);
```

From there, we could have *empty* inheritence tables.

```sql
create table gsa () inherits (gov);
create table hhs () inherits (gov);
create table nih () inherits (gov);
```

and, from there, the next level down:

```sql
create table cc () inherits (nih);
create table nccih () inherits (nih);
create table nia () inherits (nih);
```

Then, insertions happen at the **leaves**. That is, we only insert at the lowest level of the hierarchy. However, we can then query tables higher up, and get results from the entire tree.

This does two things:

1. It lets queries against a given domain happen naturally. If we want to query `nia.nih.gov`, we target that table with our query.
2. If we want to query all of `nih`, then we query the `nih` table.
3. If we want to query everything, we target `gov` (or another tld).

Given that we are going to treat these tables as build artifacts, we can always regenerate them. And, it is possible to add new tables through a migration easily; we just add a new create table statement.

(See [this article](https://medium.com/miro-engineering/sql-migrations-in-postgresql-part-1-bc38ec1cbe75) about partioning/inheritence, indexing, and migrations. It's gold.)

### declarative partitioning

Another approach is to use `PARTITION`s.

This would suggest our root table has columns we can use to drive the derivative partitions.

```sql
create table gov (
  id ...,
  domain64 BIGINT,
  host_path ...,
  tag ...,
  content ...
  partition by range(domain64)
);
```

To encode all of the TLDs, domains, and subdomains we will encounter, we'll use a `domain64` encoding. Why? It maps the entire URL space into a single, 64-bit number (or, `BIGINT`).

```
FF:FFFFFF:FFFFFF:FF
```

or

```
tld:domain:subdomain:subsub
```

This is described more in detail in [domain64.md](domain64.md).

As an example:

| tld | domain | sub |                  hex |               dec |
|-----|--------|-----|----------------------|-------------------|
| gov |    gsa |  _  |   #x0100000100000000 | 72057598332895232 |
| gov |    gsa | tts |   #x0100000100000100 | 72057598332895488 |
| gov |    gsa | api |   #x0100000100000200 | 72057598332895744 |

GSA is from the range #x0100000001000000 -> #x0100000001FFFFFF, or 72057594054705152 -> 72057594071482367 (a diff of 16777215). Nothing else can be in that range, because we're using the bitstring to partition off ranges of numbers.

Now, everything becomes bitwise operations on 64-bit integers, which will be fast everywhere... and, our semantics map well to our domain.

Partitioning to get a table with only GSA entries is

```sql
CREATE TABLE govgsa PARTITION OF gov
    FOR VALUES FROM (72057598332895232) TO (72057602627862527);
```

Or, just one subdomain in the space:

```sql
CREATE TABLE govgsatts PARTITION OF gov
    FOR VALUES FROM (72057598332895488) TO (72057598332895743);
```

or we can keep the hex representation:

```sql
CREATE TABLE govgsatts PARTITION OF gov
    FOR VALUES FROM (select x'0100000100000100') TO (select x'01000001000001FF');
```

All table operations are on the top-level table (insert, etc.), the indexes and whatnot are inherited automatically, and I can search the TLD, domain, or subdomain without difficulty---because it all becomes a question of what range the `domain64` value is in.



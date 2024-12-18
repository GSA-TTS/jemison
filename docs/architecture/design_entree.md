# entree

Entree keeps track of what is coming in the front door.

```mermaid
sequenceDiagram
  autonumber
  participant Q as Queue
  participant E as Extract
  participant DB as Postgres
  participant F as Fetch
  # ---
  Q->>E: get job
  note right of Q: host, path, hallpass?
  E->>DB: 
  activate DB
  DB->>E: OK, NO
  deactivate DB
  alt have hallpass?
  E->>F: enqueue fetch
  else is OK?
  E->>F: enqueue fetch
  E->>DB: update next_fetch
  end
```

The `entree` component decides what URLs can, or cannot, come in the door.

It helps maintain a `guestbook`. The initial design looks like

```sql
CREATE TABLE IF NOT EXISTS guestbook (
  id BIGINT generated always as identity primary key,
  scheme scheme NOT NULL,
  host BIGINT references hosts(id),
  path TEXT NOT NULL,
  content_sha1 TEXT,
  content_length INTEGER,
  content_type INTEGER references content_types(id),
  last_updated DATE,
  last_fetched DATE,
  next_fetch DATE,
  UNIQUE (host, path)
);
```

URLs enter the system via one of two ways at the moment:

1. A scheduled kickoff of a crawl, or 
2. The `walk` component

For example, if we kick off a crawl of `allthe.gov`, that URL is sent to `entree`. That URL may have a `weekly` cadence. 

* If we last crawled it on Monday, it is now Wednesday, and the `next_fetch` date is the Monday prior, we will quiety ignore the message. (Log it, etc. perhaps.)
* If we craweled on Monday, it is Wednesday, and `hallpass?` is true, it means we injected this with the intent of running it now. We do two things (and *don't* do one):
  1. Rewrite the `next_fetch` on all URLs for this host to one minute in the past.
  2. Enqueue the URL to `fetch`
  3. Do *nothing* to `next_fetch`
* If we crawled on Monday, and it is now Tuesday of the next week, we pass this on to `fetch`, and set `next_fetch` appropriately.


We pass things through `fetch`, `extract`, etc. until `serve` picks up a database. At that point, a `done` message is posted back to the `entree` queue. This lets us update the `last_fetched` for everything in the domain. (Or... fetch updates the individual rows?)

The update rules are stored in static configuration files shipped with the app. These may become dynamic at a later point.


```mermaid
flowchart TD
%% Nodes
A("Url enters"):::blue

AP("Accept"):::green
RP("Reject"):::red

AS("Accept"):::green
RS("Reject"):::red

D("Check host<br>next_fetch"):::purple

C("Check<br>Guestbook"):::orange
E("Check page<br>next_fetch"):::orange

F("Set host<br>next_fetch to<br>yesterday"):::purple

G("fetch page,<br>update last_fetch in gbook")
H("update last_fetch")

%% Edges
A -- full/no pass --> D
A -- full/pass --> F
F --> AS

A -- single/pass --> AP
A -- single/no pass --> C
C -- not in gbook --> D
C -- in gbook --> E
E -- after --> AP
E -- before --> RP

AP --> H

D -- after --> AS
D -- before --> RS

AS --> G

%% Styling
classDef green fill:#B2DFDB,stroke:#00897B,stroke-width:2px;
classDef orange fill:#FFE0B2,stroke:#FB8C00,stroke-width:2px;
classDef blue fill:#BBDEFB,stroke:#1976D2,stroke-width:2px;
classDef yellow fill:#FFF9C4,stroke:#FBC02D,stroke-width:2px;
classDef pink fill:#F8BBD0,stroke:#C2185B,stroke-width:2px;
classDef purple fill:#E1BEE7,stroke:#8E24AA,stroke-width:2px;
classDef red fill:#CC0000,stroke:#FB8C00,stroke-width:2px,color:#FFFFFF;

```

### tooling

The service uses `dbmate` to manage the `guestbook`. This is currently managed by the application directly. `dbmate` is a standalone tool, and could be extracted and scripted separately. 

It is a (largely) pure-SQL tool, and helps keep migrations expressed in a lowest-common-demoninator language. Having the app run the migrations directly is possible because `dbmate` is written in Go, and can be used as a library. However, pulling it out and running it via the `dbmate` command-line tool adds nothing in a future refactor.

## resources

* https://github.com/amacneil/dbmate
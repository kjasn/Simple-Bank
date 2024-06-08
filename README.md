# Simple-Bank

[README-zh](./README-zh.md)

## Requirements

This project requires the following tools and libraries to be installed on your system. Please ensure you have the specified versions or later.

### Tools

- **Docker**: (v24.0.9-1)

  - Installation: Follow the official Docker installation guide for your operating system.

- **PostgreSQL**: (version: 12-alpine)

  - Installation: (install with docker) `docker pull postgres:12-alpine`

- **GNU Make**: (v4.2.1)

  - Installation: Use your system's package manager. For example, on Ubuntu, you can use `sudo apt install make`.

- **Migrate**: (v4.17.0) Using to build DB with sql files

  - Installation: Use the following command to install Migrate: (official guide may occurs some mistakes: [issues#818](https://github.com/golang-migrate/migrate/issues/818#issuecomment-1270444615))
    `1. wget http://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.deb`  
    `2. sudo dpkg -i migrate.linux-amd64.deb`

- **Sqlc**: (v1.25.0) Using to generate CRUD code

  - Installation: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

- **lib/pq**: (v1.10.9) Using to provide a driver that implements postgres

  - Installation: `go get github.com/lib/pq`

- **testify** (v1.9.0) Using to check the unit test return

  - Installation: `go get github.com/stretchr/testify`

- **Gin** (v1.9.1)

  - Installation: `go get -u github.com/gin-gonic/gin`

- **Viper** (v1.18.2) Using to load configurations from files or environment variables

  - Installation: `go get github.com/spf13/viper`

- **gomock** (v1.6.0)

  - Installation: `go get github.com/golang/mock/mockgen@v1.6.0`

- **JWT** (v3.2.0)

  - Installation: `go get github.com/dgrijalva/jwt-go`

- **PASETO** (v1.0.0)

  - Installation: `go get -u github.com/o1egl/paseto`

- **dbdocs** Using for create web-based documentation.

  - Installation: `npm install -g dbdocs`

- **dbml/cli** Using for generate db schema from dbml.

  - Installation: `npm install -g @dbml/cli`

- **protobuf complier** (libprotoc 27.0-rc2)

  - Installation: `apt install -y protobuf-compiler`
    > For generate code, we need install 2 plugins, details in [gRPC documentation](https://grpc.io/docs/languages/go/quickstart/).
    > When it occurs some errors, like Timestamp not found, [here](https://stackoverflow.com/questions/56031098/protobuf-timestamp-not-found) is a way to solve.

- **statik**

  - Installation: `go get github.com/rakyll/statik`

- **zerolog**

  - Installation: `go get -u github.com/rs/zerolog/log`

- **asynq**

  - Installation: `go get -u github.com/hibiken/asynq`

## Deal With Concurrency And Deadlock

First, we add `FOR UPDATE` to function `GetAccount`:

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR UPDATE;
```

Then `make sqlc`, we get a new function `GetAccountForUpdate`, it ensures the correctness of concurrent data through **MUTUAL EXCLUSION**. But we get another problem: `Error: deadlock detected` (when the number of concurrency is enough. Because I only set 2 concurrency goroutine, it passed).

Second, to deal with deadlock. We can print logs while running and find what causes deadlock. Thus we can cleanly see which option causes deadlock.

The following query may be helpful to see what processes are blocking SQL statements (these only find row-level locks, not object-level locks). Copy from [wiki](https://wiki.postgresql.org/wiki/Lock_Monitoring)

```sql
SELECT blocked_locks.pid     AS blocked_pid,
        blocked_activity.usename  AS blocked_user,
        blocking_locks.pid     AS blocking_pid,
        blocking_activity.usename AS blocking_user,
        blocked_activity.query    AS blocked_statement,
        blocking_activity.query   AS current_statement_in_blocking_process
FROM  pg_catalog.pg_locks         blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks         blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid

JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

The following query allow us to list all the locks in our database.

```sql
SELECT
        a.application_name,
        l.relation::regclass,
        l.transactionid,
        l.mode,
				l.locktype,
        l.GRANTED,
        a.usename,
        a.query,
        a.pid
FROM pg_stat_activity a
JOIN pg_locks l ON l.pid = a.pid
WHERE a.application_name = 'psql'
ORDER BY a.pid;
```

We may find a select option from accounts table needs to get lock from other transaction that runs insert option on transfers table. Back to schema:

```sql
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
```

The only connection between transfers table and accounts table is the **FOREIGN KEY CONSTRAINT**. Each update of accounts table may cause it acquires lock from transfers table. Now we know our option will not update account id, so we should tell postgres12 this. Then we change to `FOR NO KEY UPDATE`, and `make sqlc`

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;
```

However, the above ways to deal with deadlock may still occurs some mistakes.
For example, there are 2 transactions, one transfers money from account1 to account2, and the other transfers from account2 back to account1.

The order of update account's balance:

- Transfer1: account1 - amount --> account2 + amount
- Transfer2: account2 - amount --> account1 + amount

So, before each of them commit, they hold a exclusive lock which blocks the other.

The best way is to avoid deadlock by making sure that the transfers are processed **in an consistent order**. Like we can enable each transfer update account with smaller ID first.

## About Transaction Isolation Level

The following table is copy from [postgressql document](https://www.postgresql.org/docs/current/transaction-iso.html)

| Isolation Level  |         Dirty Read         | Nonrepeatable Read |        Phantom Read        | Serialization Anomaly |
| :--------------: | :------------------------: | :----------------: | :------------------------: | :-------------------: |
| Read uncommitted | **Allowed, but not in PG** |      Possible      |          Possible          |       Possible        |
|  Read committed  |        Not possible        |      Possible      |          Possible          |       Possible        |
| Repeatable read  |        Not possible        |    Not possible    | **Allowed, but not in PG** |       Possible        |
|   Serializable   |        Not possible        |    Not possible    |        Not possible        |     Not possible      |

Read uncommitted is the **SAME** as read committed(default level) **in postgres**. lBasically, there are 3 isolation levels in postgres.

Postgres uses **dependencies checking mechanism** to prevent the **serialization anomaly**, while MySQL uses **locking mechanism**.

[More details in my blog](https://kjasn.github.io/2024/03/18/Transaction-isolation-level-of-DB/)...

## UnitTest in Mock

Mock store is separated for each test case.

\#TODO

## Add Users Table

Instead of resetting the old database and create a new one then migrate data to it, we are supposed to update it with a new edition.

We set users table as follows:

- Username as primary key.
- Each account must links to a user via foreign key -- username.
- Each user can have many accounts with **different** currency.
- Each email address can only bind one user, it means email field is unique.

## Encrypt Password

Basically, we do not save plain password text, instead we prefer to save encrypted password. Here we use `bcrypt` to generate hashed password.

In this encrypt algorithm, we use a **random salt** and a cost(the iterate times) to encrypt the provided password:

1. Although the provided passwords are the same, it can generates different hash value.

2. For comparing and checking password, it use the cost and the same salt from the hash value to encrypt the provided password, then check if the new hash value equals the provided one.

## Authentication & Authorization

We can make authentication via a specific middleware, and register is for a router group which contain all api that need be authorized before pass to call real handlers.

For authorization, it is API specific.

\#TODO

## User Session Management

We should not use JWT or PASETO as a token-based authentication for **long session**.
Because of the **stateless design**, those tokens are **not stored in database**, when they are leaked, there is no way to revoke them. Therefore we must **set their lifetime short enough**(about 10-15 min). But if we only use access tokens, when they are expired, users need to frequently login to get new tokens. It is a terrible user experience.

Now we can additionally use a refresh token to maintain a **stateful session** on the server, and the client can use it with a long valid duration to request a new access token when it is expired. The refresh token will be **stored in a session table in the database**, so we can revoke it easily and its lifetime can be much longer(such as several days).

## gRPC And HTTP Serve

gRPC is famous for its high performance which is very suitable for microservice and mobile application, so we can replace HTTP JSON APIs with it. But sometimes we might still provide normal HTTP JSON APIs to client.Thus we need an idea to serve both HTTP JSON APIs and gRPC, and that is gRPC gateway.

\#TODO

Since **the first server we run will block the second one**, so we need run them in different go routine.

## Asynchronous Processing

For synchronous APIs, when clients send request to the server, the request **must be processed immediately by the server and the result will be return synchronously**. But sometimes when the request can not to be processed immediately, maybe it **takes long time to complete** and we do not want to force clients to wait, OR maybe we just want to **schedule it to be execute in the future**. Thus we need a mechanism to process some kinds of tasks asynchronously.

Firstly, we may think about using go routine to process in the background because it is simple to implement. But the drawback is the task live inside in the process's memory. If the server goes down, the unprocessed tasks may be lost.

Then using message queue will be a better design. Redis is an efficient message queue, which stores its data both in memory and persistent storage. With its high-concurrency and reliability, our tasks will not be lost.

# Simple-Bank

## Requirements

This project requires the following tools and libraries to be installed on your system. Please ensure you have the specified versions or later.

### Tools

- **Docker**: Version 24.0.9-1
    - Installation: Follow the official Docker installation guide for your operating system.

- **PostgreSQL**: Version 12-alpine
    - Installation: (install with docker) `docker pull postgres:12-alpine`

- **GNU Make**: Version 4.2.1
    - Installation: Use your system's package manager. For example, on Ubuntu, you can use `sudo apt install make`.

- **Migrate**: Version 4.17.0   using to build DB with sql files
    - Installation: Use the following command to install Migrate: (official guide may occurs some mistakes: [issues#818](https://github.com/golang-migrate/migrate/issues/818#issuecomment-1270444615)) 
    `1. wget http://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.deb`         
    `2. sudo dpkg -i migrate.linux-amd64.deb`

- **Sqlc**: Version 1.25.0      using to generate CRUD code 
    - Installation: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`     
 
- **lib/pq**: Version 1.10.9    using to provide a driver that implements postgres
    - Installation: `go get github.com/lib/pq`

- **testify** Version 1.9.0     using to check the unit test return
    - Installation: `go get github.com/stretchr/testify`


## TODO

1. `deleteXxx` function for entries and transfers
2. Search entries or transfers by account id AND how to automatically generate the unit tests.
3. deal with the deadlock


## Deal with concurrency and deadlock

First, we add `FOR UPDATE` to function `GetAccount`:
```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts 
WHERE id = $1 LIMIT 1
FOR UPDATE;
``` 
Then `make sqlc`, we get a new function `GetAccountForUpdate`, it ensures the correctness of concurrent data through **MUTUAL EXCLUSION**. But we another problem: `Error: deadlock detected` (when the number of concurrency is enough. Because I have set 2 concurrency goroutine, it passed).

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
SELECT a.datname,
        l.relation::regclass,
        l.transactionid,
        l.mode,
        l.GRANTED,
        a.usename,
        a.query,
        a.query_start,
        age(now(), a.query_start) AS "age",
        a.pid
FROM pg_stat_activity a
JOIN pg_locks l ON l.pid = a.pid
ORDER BY a.query_start;
```

We may find a select option from accounts table needs to get lock from other transaction that runs insert option on transfers table.  Back to schema:

```sql
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
```

The only connection between transfers table and accounts table is the **FOREIGN KEY CONSTRAINT**. Each update of accounts table may cause it acquire lock from transfers table. Now we know our option will not update account id, so we should tell postgres12 this. Then we change to `FOR NO KEY UPDATE`, and `make sqlc`

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts 
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;
```




# Simple-Bank

## 需求

该项目需要在您的系统上安装以下工具和库。请确保您拥有指定版本或更高版本。

### 工具

- **Docker**: (v24.0.9-1)

  - 安装：按照官方的 Docker 安装指南为您的操作系统进行安装。

- **PostgreSQL**: （版本：12-alpine）

  - 安装：（使用 docker 安装）`docker pull postgres:12-alpine`

- **GNU Make**: (v4.2.1)

  - 安装：使用您系统的包管理器。例如，在 Ubuntu 上，您可以使用 `sudo apt install make`。

- **Migrate**: (v4.17.0) 用于使用 sql 文件构建数据库

  - 安装：使用以下命令安装 Migrate：（官方指南可能出现一些错误：[问题#818](https://github.com/golang-migrate/migrate/issues/818#issuecomment-1270444615))
    `1. wget http://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.deb`  
    `2. sudo dpkg -i migrate.linux-amd64.deb`

- **Sqlc**: (v1.25.0) 用于生成 CRUD 代码

  - 安装：`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

- **lib/pq**: (v1.10.9) 用于提供实现 postgres 的驱动

  - 安装：`go get github.com/lib/pq`

- **testify** (v1.9.0) 用于检查单元测试返回

  - 安装：`go get github.com/stretchr/testify`

- **Gin** (v1.9.1)

  - 安装：`go get -u github.com/gin-gonic/gin`

- **Viper** (v1.18.2) 用于从文件或环境变量加载配置

  - 安装：`go get github.com/spf13/viper`

- **gomock** (v1.6.0)

  - 安装：`go get github.com/golang/mock/mockgen@v1.6.0`

- **JWT** (v3.2.0)

  - 安装：`go get github.com/dgrijalva/jwt-go`

- **PASETO** (v1.0.0)

  - 安装：`go get -u github.com/o1egl/paseto`

- **dbdocs** 用于创建基于 web 的文档。

  - 安装：`npm install -g dbdocs`

- **dbml/cli** 用于从 dbml 生成数据库架构。

  - 安装：`npm install -g @dbml/cli`

- **protobuf 编译器** (libprotoc 27.0-rc2)

  - 安装：`apt install -y protobuf-compiler`
    > 为了生成代码，我们需要安装 2 个插件，详细信息在 [gRPC 文档](https://grpc.io/docs/languages/go/quickstart/)。
    > 当出现一些错误时，比如 Timestamp 找不到，[这里](https://stackoverflow.com/questions/56031098/protobuf-timestamp-not-found) 是一种解决方法。

- **statik**
  - 安装：`go get github.com/rakyll/statik`

## 处理并发和死锁

首先，我们在 `GetAccount` 函数中添加 `FOR UPDATE`：

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR UPDATE;
```

然后 `make sqlc`，我们得到一个新的函数 `GetAccountForUpdate`，它通过 **互斥** 确保并发数据的正确性。但我们遇到了另一个问题：`Error: deadlock detected`（当并发数足够多时。因为我只设置了 2 个并发 goroutine，它通过了）。

接着，为了解决死锁问题。我们可以在运行时打印日志并找出导致死锁的原因。这样我们可以清楚地看到哪个选项导致死锁。

以下查询可能有助于查看哪些进程正在阻止 SQL 语句（这些只找到行级锁，而不是对象级锁）。从 [wiki](https://wiki.postgresql.org/wiki/Lock_Monitoring) 复制

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

以下查询允许我们列出我们数据库中的所有锁。

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

我们可能发现从 accounts 表中选择一个选项需要从在 transfers 表上运行 insert 选项的其他事务中获取锁。回到架构：

```sql
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
```

transfers 表和 accounts 表之间的唯一连接是 **外键约束**。accounts 表的每次更新都可能导致它从 transfers 表获取锁。现在我们知道我们的选项不会更新账户 id，所以我们应该告诉 postgres12。然后我们改为 `FOR NO KEY UPDATE`，并 `make sqlc`

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;
```

然而，上述处理死锁的方法仍然可能出现一些错误。
例如，有 2 个事务，一个从 account1 向 account2 转账，另一个从 account2 转回 account1。

更新账户余额的顺序：

- Transfer1: account1 - amount --> account2 + amount
- Transfer2: account2 - amount --> account1 + amount

所以，在它们各自提交之前，它们都持有一个独占锁，这阻止了另一个。

最好的方法是通过确保事务以 **一致的顺序** 处理来避免死锁。比如我们可以允许每个转账先更新 ID 较小的账户。

## 关于事务隔离级别

以下表格从 [postgressql 文档](https://www.postgresql.org/docs/current/transaction-iso.html) 复制

| 隔离级别 |           脏读           | 可重复读 |           幻读           | 串行化异常 |
| :------: | :----------------------: | :------: | :----------------------: | :--------: |
| 读未提交 | **允许，但在 PG 中不是** |   可能   |           可能           |    可能    |
| 读已提交 |          不可能          |   可能   |           可能           |    可能    |
| 可重复读 |          不可能          |  不可能  | **允许，但在 PG 中不是** |    可能    |
| 可串行化 |          不可能          |  不可能  |          不可能          |   不可能   |

在 postgres 中，读未提交是 **与读已提交（默认级别）相同** 的。基本上，在 postgres 中有 3 个隔离级别。

Postgres 使用 **依赖性检查机制** 来防止 **串行化异常**，而 MySQL 使用 **锁定机制**。

[更多详情在我的博客](https://kjasn.github.io/2024/03/18/Transaction-isolation-level-of-DB/)...

## Mock 中的单元测试

Mock store 对于每个测试用例是分开的。

\#TODO

## 添加用户表

我们不应该重置旧数据库并创建一个新的，然后将数据迁移到新数据库中，我们应该用新版本更新它。

我们设置用户表如下：

- 用户名作为主键。
- 每个账户必须通过外键——用户名与用户关联。
- 每个用户可以拥有多个具有**不同**货币的账户。
- 每个电子邮件地址只能绑定一个用户，这意味着电子邮件字段是唯一的。

## 加密密码

一般我们不保存明文密码，我们更倾向于保存加密密码。在这里，我们使用 `bcrypt` 生成哈希密码。

在这个加密算法中，我们使用一个**随机盐值**和一个成本（迭代次数）来加密提供的密码：

1. 即使提供的密码相同，它也可以生成不同的哈希值。

2. 为了比较和检查密码，它使用哈希值中的成本和相同的盐来加密提供的密码，然后检查新的哈希值是否等于提供的哈希值。

## 认证与授权

我们可以通过特定的中间件进行认证，注册是针对包含所有需要在调用真实处理程序之前进行授权的 API 的路由器组。

对于授权，它是特定于 API 的。

\#TODO

## 用户会话管理

我们不应该使用 JWT 或 PASETO 作为**长会话**的基于令牌的认证。
由于**无状态设计**，这些令牌**不存储在数据库中**，当它们被泄露时，没有办法撤销它们。因此，我们必须**设置它们足够短的生命周期**（大约 10-15 分钟）。但是，如果我们只使用访问令牌，当它们过期时，用户需要频繁登录以获取新的令牌。这是一个可怕的用户体验。

现在，我们可以额外使用一个刷新令牌来在服务器上维护一个**有状态的会话**，客户端可以使用它在长时间内有效，以便在过期时请求一个新的访问令牌。刷新令牌将**存储在数据库中的会话表**中，所以我们可以轻松地撤销它，它的生命周期可以更长（例如几天）。

## gRPC 和 HTTP 服务

gRPC 以其高性能而闻名，非常适合微服务和移动应用，所以我们可以用它来替代 HTTP JSON API。但有时我们可能仍然需要向客户端提供正常的 HTTP JSON API。因此，我们需要一个想法来同时提供 HTTP JSON API 和 gRPC，那就是 gRPC 网关。

\#TODO

由于**我们运行的第一个服务器将阻塞第二个服务器**，所以我们需要在不同的 goroutine 中运行它们。

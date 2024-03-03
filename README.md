# Simple-Bank

## Requirements

```bash
postgres version 12-alpine
Docker version 24.0.9-1
GNU Make 4.2.1
migrate version 4.17.0
```

```bash
# migrate schema files to databse
migrate -path db/migration -database "postgres://root:8520@localhost:5432/simple_bank?sslmode=disable" -verbose up
```
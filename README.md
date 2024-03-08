# Simple-Bank

## Requirements

This project requires the following tools and libraries to be installed on your system. Please ensure you have the specified versions or later.

### Tools

- **PostgreSQL**: Version 12-alpine
    - Installation: Follow the official PostgreSQL installation guide for your operating system.

- **Docker**: Version 24.0.9-1
    - Installation: Follow the official Docker installation guide for your operating system.

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
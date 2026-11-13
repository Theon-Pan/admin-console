# Admin Console

Admin Console is a Go web application that uses PostgreSQL for persistent data. It includes an HTTP server, background workers, user and web-session support, and a web UI.

## Requirements

- Go 1.27 or later
- PostgreSQL

## Getting started

Create a PostgreSQL database for the application, then configure its connection string:

```sh
createdb admin-console
export PGPASSWORD=your-postgres-password
export DATABASE_URL_AC='user=postgres host=localhost port=5432 dbname=admin-console sslmode=disable'
```

Run database migrations and create an administrator account:

```sh
go run . -migrate
go run . -create-admin
```

Start the application:

```sh
go run .
```

By default, the HTTP server listens on `127.0.0.1:8081`. The database connection can also be set in a configuration file and loaded with `-config-file <path>`; environment variables take precedence over values in the file.

## Configuration

Configuration is read from environment variables, with optional settings loaded from a file passed to `-config-file`. Common settings include:

| Variable | Description | Default |
| --- | --- | --- |
| `DATABASE_URL_AC` | PostgreSQL connection string | No usable connection string is provided; set this before running |
| `LISTEN_ADDR` | HTTP listen address (comma-separated for multiple addresses) | `127.0.0.1:8081` |
| `LOG_FILE` | Log destination (`stderr`, `stdout`, or a file path) | `stderr` |
| `LOG_LEVEL` | Log level: `debug`, `info`, `warning`, or `error` | `info` |
| `WORKER_POOL_SIZE` | Number of background workers | `16` |

See `conf/admin-console.dev.conf` for an example of additional configuration options. Set `DATABASE_URL_AC` to a valid connection string before using that file.

## Development

### Debugging in VS Code

Install the Go extension (`golang.go`) and use **Go: Install/Update Tools** from
the Command Palette to install `dlv` (Delve).

Start PostgreSQL and Redis/Valkey, and configure their connection settings in
`conf/admin-console.dev.conf` or environment variables. Environment variables
override the file; do not put credentials in the shared launch configuration.
Run migrations and create an administrator account first if needed.

Open the repository root in VS Code, set a breakpoint, select **Debug Admin
Console** in **Run and Debug**, and press **F5**. The launch profile uses the
development configuration with debug logging in the integrated terminal.
Send an HTTP request to the configured listen address to hit handler breakpoints.
Use **Shift+F5** to stop debugging.

Run the Go test suite and build the application with:

```sh
go test ./...
go build -o admin-console .
```

The React UI source is in `react-ui/`; its README and `package.json` describe its development commands.

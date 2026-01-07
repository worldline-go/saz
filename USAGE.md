# SAZ Usage Guide

This document provides detailed usage instructions for SAZ (Structured Access Zone).

## Quick Start

```sh
# Create PostgreSQL environment
make env

# Run the service
make run
```

Open [http://localhost:8080](http://localhost:8080) to access the web UI.

### Docker

```sh
docker run -p 8080:8080 ghcr.io/worldline-go/saz:latest
```

## Configuration

Configuration can be provided via YAML file, environment variables, Consul, or Vault.

### Example Configuration (`saz.yaml`)

```yaml
# Server configuration
server:
  host: ""              # Bind host (default: all interfaces)
  port: "8080"          # Server port
  base_path: ""         # Base URL path (e.g., "/saz")
  private_token: ""     # Optional authentication token (Private-Token header)

# Log level: debug, info, warn, error
log_level: "info"

# Database connectionsS
database:
  # Custom named databases
  my-postgres-demo:
    db_datasource: "postgres://postgres@localhost:5432/postgres?sslmode=disable"
    db_type: "pgx"

  my-oracle:
    db_datasource: "user/password@localhost:1521/orcl"
    db_type: "godror"

# Store configuration (for saving notebooks)
store:
  postgres:
    table_prefix: "saz_"
    db_datasource: "postgres://postgres@localhost:5432/postgres?sslmode=disable"
    db_type: "pgx"
    migrate:
      db_datasource: "postgres://postgres@localhost:5432/postgres?sslmode=disable"
      db_schema: "public"
```

### Supported Database Types

| Type        | Driver     | Description          |
| ----------- | ---------- | -------------------- |
| `pgx`       | pgx        | PostgreSQL           |
| `godror`    | godror     | Oracle               |
| `sqlite3`   | sqlite3    | SQLite3              |
| `sqlserver` | go-mssqldb | Microsoft SQL Server |
| `mysql`     | mysql      | MySQL                |
| `odbc`      | odbc       | ODBC connections     |

## Web UI Features

### Notebooks

SAZ provides a notebook-style interface for organizing and executing SQL queries:

- **Create notebooks** to organize related queries
- **Add cells** with SQL queries to each notebook
- **Execute cells** individually or run entire notebooks
- **Export results** to CSV

### Cell Configuration

Each cell in a notebook supports:

| Option      | Description                               |
| ----------- | ----------------------------------------- |
| Database    | Select which configured database to query |
| SQL Content | The SQL query to execute                  |
| Description | Human-readable description                |
| Path        | Unique identifier for cell references     |
| Limit       | Maximum rows to return                    |
| Result      | Toggle whether to return results          |
| Template    | Enable Go templating in SQL               |
| Enabled     | Include/exclude from notebook execution   |

### Template Support

Check all template functions in [mugo reference page](https://rytsh.github.io/mugo/functions/reference.html)  
Enable template mode to use Go templates in SQL queries:

```sql
{{$v := "postgres" }}
SELECT * FROM pg_catalog.pg_tables where tableowner = '{{ $v }}';
```

Cells can reference results from other cells by their path name.

### Dependency Management

Create a cell that other cells depend on. Dependent cells will only execute if the parent cell succeeds.

This is an example of a parent cell and lets call path-name to `tables`;  
Also you need to enable `output` and give `0` limit to get all results.

```sql
SELECT * FROM pg_catalog.pg_tables;
```

In the second cell, we can reference the first cell and use its results; open dependent mode and add `tables` in there.

Than in the SQL content, we can use with `.cells.tables` but this is an array of map results, so we can access the first row with index `0` and its `schemaname` column like this:

```sql
SELECT * FROM pg_catalog.pg_tables where schemaname = '{{ (index .cells.tables 0).schemaname }}';
```

## REST API

### Endpoints

| Method     | Endpoint                    | Description                           |
| ---------- | --------------------------- | ------------------------------------- |
| `GET`      | `/api/v1/info`              | Get service info (databases, version) |
| `POST`     | `/api/v1/run`               | Execute a query cell                  |
| `POST/GET` | `/api/v1/run/{note}`        | Execute all cells in a notebook       |
| `POST/GET` | `/api/v1/run/{note}/{cell}` | Execute a specific cell in a notebook |
| `GET`      | `/api/v1/notes`             | List all notebooks                    |
| `GET`      | `/api/v1/notes/{id}`        | Get a notebook by ID                  |
| `PUT`      | `/api/v1/notes/{id}`        | Create/update a notebook              |
| `DELETE`   | `/api/v1/notes/{id}`        | Delete a notebook                     |
| `POST`     | `/api/v1/render`            | Render a Go template                  |

### Call note or cell with POST data

You can pass data to a notebook or cell using POST requests. The data will be available in the template context.  
If `private_token` is set in the configuration, include it in the `Private-Token` header.

```sh
curl -X POST http://localhost:8080/api/v1/run/my_notebook -d '{"status":"active"}'
```

In the SQL cell template, you can access the `status` variable:

```sql
SELECT * FROM users WHERE status = '{{ .data.status }}';
```

You can also send example `status` data with GET requests by appending query parameters:

```sh
curl "http://localhost:8080/api/v1/run/my_notebook?status=active"
```

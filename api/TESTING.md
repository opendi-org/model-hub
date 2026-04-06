# API Testing Notes

## Why tests can look green when DB tests did not run

Some `api` tests are integration tests that require Postgres.  
When Postgres is unavailable or credentials are wrong, those tests call `t.Skip(...)` and `go test` still exits `0`.

The skip message looks like:

- `database not available: ... (set DB_* or TEST_DSN to run)`

If you see that message, integration coverage did not run.

## Recommended local commands

### Fast local run (may skip DB-backed integration tests)

```bash
go test ./...
```

### Full DB-backed run with Docker Compose

From repo root:

```bash
docker compose -p openditest -f compose.yaml --env-file .env up -d db
docker compose -p openditest -f compose.yaml --env-file .env run --build --rm -T api go test -p 1 -v ./...
docker compose -p openditest -f compose.yaml --env-file .env down
```

Notes:

- Use `-p openditest` to keep test containers isolated from your normal stack.
- Use `go test -p 1` for the Docker-backed integration run to avoid cross-package migration races.
- `docker compose ... down` (without `-v`) keeps DB volumes intact.

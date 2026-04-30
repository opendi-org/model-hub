# API Testing Notes

## Why tests can look green when DB tests did not run

Some `api` tests are integration tests that require Postgres.  
When Postgres is unavailable or credentials are wrong, those tests call `t.Skip(...)` and `go test` still exits `0`.

The skip message looks like:

- `database not available: ... (set DB_* or TEST_DSN to run)`

If you see that message, integration coverage did not run.

Fixture-dependent handler tests also rely on files under
`api/cdm-json-schema/examples`. If that directory is missing (for example, a
checkout without submodule content), those tests intentionally skip instead of
failing.

## Recommended local commands

### Fast local run (may skip DB-backed integration tests)

```bash
go test -v ./...
```

To explicitly see whether integration tests were skipped:

```bash
go test -v ./... | rg "SKIP|database not available"
```

### Full DB-backed run with Docker Compose

From repo root:

```bash
docker compose -p openditest -f compose.dev.yaml up -d db
docker compose -p openditest -f compose.dev.yaml run --build --rm -T api go test -p 1 -v ./...
docker compose -p openditest -f compose.dev.yaml down
```

Notes:

- Use `-p openditest` to keep test containers isolated from your normal stack.
- Use `go test -p 1` for the Docker-backed integration run to avoid cross-package migration races.
- `docker compose ... down` (without `-v`) keeps DB volumes intact.
- If fixture-based tests are skipped, ensure submodule content is checked out:
  - `git submodule update --init --recursive`

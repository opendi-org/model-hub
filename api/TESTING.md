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

## Local Testing Prerequisite: start the test database

Tests expect a db accessible on `localhost:5432` with these vars:
- `DB_NAME`: `modelhub_test`
- `DB_USERNAME`: `postgres`
- `DB_PASSWORD`: `postgres`

Use the testing compose `compose.test.yaml` to set up a dedicated testing db. It hardcodes the test suite's
defaults on top of `compose.dev.yaml`.

```bash
docker compose -p openditest -f compose.dev.yaml -f compose.test.yaml up -d db
```

Start the container and you can run `go test` or use VS Code's Go testing extensions or whatever else.

Notes:

- Use `-p openditest` to keep this container and its volume
  (`openditest_db_data`) isolated from your normal dev stack's.
- It still binds host port `5432`, same as the normal dev stack's `db`, so
  **the two can't run at the same time**. Bring this one down
  (`docker compose -p openditest -f compose.dev.yaml -f compose.test.yaml down`)
  before starting your normal `docker compose up`, and vice versa.
- `docker compose ... down` (without `-v`) keeps DB volumes intact.

## Recommended local commands

### Fast local run

With the test database above running:

```bash
go test -v ./...
```

To explicitly see whether integration tests were skipped (e.g. because you
forgot to start the test database):

```bash
go test -v ./... | rg "SKIP|database not available"
```

### Full DB-backed run with Docker Compose

From repo root, with the test database above running:

```bash
docker compose -p openditest -f compose.dev.yaml -f compose.test.yaml run --build --rm -T api go test -p 1 -v ./...
```

Notes:

- Use `go test -p 1` for the Docker-backed integration run to avoid
  cross-package migration races.
- If fixture-based tests are skipped, ensure submodule content is checked out:
  - `git submodule update --init --recursive`
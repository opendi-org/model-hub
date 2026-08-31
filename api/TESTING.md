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

### Check certs for `db` and `api`

Spin up a test production with just these services to check the connection.  
From repo root:  
```bash
docker compose -p openditest-prod-db-api-certs -f compose.prod.yaml up db api --build
```

Expect output like this:
```
<CUT FOR LENGTH, THESE ARE THE LAST FEW LINES>
db-1  | 2026-08-31 18:29:42.606 UTC [1] LOG:  database system is ready to accept connections
Container openditest-prod-db-api-certs-db-1 Healthy 
api-1  | 2026/08/31 18:29:46 running database migrations...
api-1  | 2026/08/31 18:29:46 model-hub listening on 0.0.0.0:8080 (dev=false)
```

Test the connection:

First, force `api` to open/use a connection via curl to an open endpoint:  
```bash
docker run --rm --network openditest-prod-db-api-certs_backend curlimages/curl -s http://api:8080/v0/search?q=test
```

For a new Compose project with no repositories matching the query "test", expect output like  this:
```json
{"repositories":[],"total":0}
```

Next, look for the connection among `db`'s current connections:
```bash
docker compose -p openditest-prod-db-api-certs -f compose.prod.yaml exec db psql -U postgres -d model_hub -c "SELECT pid, usename, ssl, client_addr FROM pg_stat_ssl JOIN pg_stat_activity USING (pid);"
```
**NOTE:** This `psql` command must run shortly after the prior `curl` command, before the `api` connection ages out of the pool.

Expect output like this:
```
 pid | usename  | ssl | client_addr 
-----+----------+-----+-------------
  85 | postgres | t   | 172.22.0.3
 220 | postgres | f   | 
(2 rows)
```

Here, row `pid=85` is the connection from `api`. `ssl=t`, so the connection is encrypted. The second process (`pid=220`) with no `client_addr` is the `psql` session's local socket connection.

If you see errors during database initialization or at any other point in the process of running these commands, check the Prerequisites section of the Deployment Guide for info on generating certs. You'll likely need to generate new certs with `db/generate-cert.sh`.

To clean up after:
```bash
docker compose -p openditest-prod-db-api-certs -f compose.prod.yaml down -v
```
**(this will also delete the database volume and any data stored there)**
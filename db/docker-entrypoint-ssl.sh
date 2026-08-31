#!/bin/sh
# Wrapper entrypoint for the `db` service (production).
#
# Postgres refuses to start if its SSL private key is group/world-readable.
# Getting permissions to transfer on bind-mount is unreliable depending on
# host OS. See comments in db/generate-cert.sh.
# 
# The postgres container starts as root though, so this wrapper handles copying
# certs into container-local paths and fixing perms before handing things
# back to the original entrypoint, which will then change user to `postgres`.
set -e

CERT_DIR=/var/lib/postgresql/certs
mkdir -p "$CERT_DIR"

cp /run/certs-ro/server.crt "$CERT_DIR/server.crt"
cp /run/certs-ro/server.key "$CERT_DIR/server.key"

chown postgres:postgres "$CERT_DIR/server.crt" "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"
chmod 600 "$CERT_DIR/server.key"

exec docker-entrypoint.sh "$@"

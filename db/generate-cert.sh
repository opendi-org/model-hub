#!/bin/sh
# Generates a self-signed cert/key pair for the `db` service.
#
# The cert's common name must be "db" to match the Docker Compose service
# name the api container connects to.
# 
# As long as the compose service stays named "db", this means the same
# cert/key pair works for local dev testing and for actual prod.
#
# This script tries chmod but on Windows this likely won't do anything.
# The db service's entrypoint wrapper will enforce correct ownership/perms
#
# MSYS_NO_PATHCONV=1 fixes a git bash / windows issue with "/CN=db".
# (Shouldn't affect Linux/macOS)
#
# That same variable disables MSYS's (otherwise useful) translation of real
# file-path args into Windows paths, so -keyout/-out can't be given absolute
# paths here without breaking the same way -subj did. This cd's into OUT_DIR and
# uses bare relative filenames instead since MSYS_NO_PATHCONV=1 leaves those alone.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="$SCRIPT_DIR/../certs/db"
mkdir -p "$OUT_DIR"
cd "$OUT_DIR"

MSYS_NO_PATHCONV=1 openssl req -new -x509 -days 3650 -nodes \
  -subj "/CN=db" \
  -keyout server.key \
  -out server.crt

chmod 600 server.key 2>/dev/null || true

echo ""
echo "Generated cert/key in $OUT_DIR"
echo "Set these in .env:"
echo "  DB_SSL_CRT_PATH=$OUT_DIR/server.crt"
echo "  DB_SSL_KEY_PATH=$OUT_DIR/server.key"

#!/usr/bin/env sh

# Start gaia
exec /app/gaia-linux-amd64 --port=${GAIA_PORT} --homepath=${GAIA_HOMEPATH} --worker=${GAIA_WORKER}
FROM alpine:3.8

# Gaia internal port and data path.
ENV GAIA_PORT=8080 \
    GAIA_HOMEPATH=/data

# Directory for the binary
WORKDIR /app

# Copy gaia binary into docker image
COPY gaia-linux-amd64 /app

# Fix permissions and install g++, make and pkg-config
RUN chmod +x ./gaia-linux-amd64 && \
    apk add --no-cache --virtual g++ make pkg-config

# Set homepath as volume
VOLUME [ "${GAIA_HOMEPATH}" ]

# Expose port
EXPOSE ${GAIA_PORT}

# Copy entry point script
COPY docker-entrypoint.sh /usr/local/bin/

# Start gaia
ENTRYPOINT [ "docker-entrypoint.sh" ]

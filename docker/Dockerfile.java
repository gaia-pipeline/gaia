FROM maven:3-jdk-8

# Version and other variables which can be changed.
ENV GAIA_PORT=8080 \
    GAIA_WORKER=2 \
    GAIA_HOMEPATH=/data

# Directory for the binary
WORKDIR /app

# Copy gaia binary into docker image
COPY gaia-linux-amd64 /app

# Fix permissions and install git
RUN chmod +x ./gaia-linux-amd64

# Set homepath as volume
VOLUME [ "${GAIA_HOMEPATH}" ]

# Expose port
EXPOSE ${GAIA_PORT}

# Copy entry point script
COPY docker-entrypoint.sh /usr/local/bin/

# Start gaia
ENTRYPOINT [ "docker-entrypoint.sh" ]

FROM maven:3.6.3-openjdk-11

# Version and other variables which can be changed.
ENV GAIA_PORT=8080 \
    GAIA_HOME_PATH=/data

# Directory for the binary
WORKDIR /app

# Copy gaia binary into docker image
COPY gaia-linux-amd64 /app

# Fix permissions & setup known hosts file for ssh agent.
RUN chmod +x ./gaia-linux-amd64 \
    && mkdir -p /root/.ssh \
    && touch /root/.ssh/known_hosts \
    && chmod 600 /root/.ssh

# Set homepath as volume
VOLUME [ "${GAIA_HOME_PATH}" ]

# Expose port
EXPOSE ${GAIA_PORT}

# Copy entry point script
COPY docker/docker-entrypoint.sh /usr/local/bin/

# Start gaia
ENTRYPOINT [ "docker-entrypoint.sh" ]

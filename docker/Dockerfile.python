FROM python:2.7-alpine3.8

# Version and other variables which can be changed.
ENV GAIA_PORT=8080 \
    GAIA_WORKER=2 \
    GAIA_HOMEPATH=/data

# install additional deps
RUN set -ex; \
	apk add --no-cache build-base python-dev \
    && pip install virtualenv grpcio

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
VOLUME [ "${GAIA_HOMEPATH}" ]

# Expose port
EXPOSE ${GAIA_PORT}

# Copy entry point script
COPY docker/docker-entrypoint.sh /usr/local/bin/

# Start gaia
ENTRYPOINT [ "docker-entrypoint.sh" ]

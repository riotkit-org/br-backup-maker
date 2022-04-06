ARG BR_PGBR_VERSION
ARG BR_PGBR_DEFAULT_PG=12.0
FROM ghcr.io/riotkit-org/pgbr:${BR_PGBR_VERSION}-pg${BR_PGBR_DEFAULT_PG} as pgbr

# main image
FROM debian:11.2-slim

COPY --from=pgbr /usr/bin/pgbr /usr/bin/pgbr-${BR_PGBR_DEFAULT_PG}
RUN ln -s /usr/bin/pgbr-${BR_PGBR_DEFAULT_PG} /usr/bin/pgbr

# basic tooling for running backup jobs on Kubernetes
RUN apt-get update -y && apt-get install -y git bash curl wget gpg patchelf mysql-client && apt-get clean

# Backup Maker
ADD .build/backup-maker /usr/bin/backup-maker
RUN chmod +x /usr/bin/backup-maker

# non-root user by default
USER 1001

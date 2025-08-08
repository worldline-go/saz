ARG DB_IMAGE=ghcr.io/rytsh/dock/db/multi:debian-12.10-slim

FROM ${DB_IMAGE}

RUN apt-get update && apt-get install --no-install-recommends -y \
    netcat-openbsd ca-certificates curl iputils-ping dnsutils

COPY saz /usr/local/bin/saz

ENTRYPOINT ["/usr/local/bin/saz"]

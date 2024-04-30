# syntax=docker/dockerfile:1

FROM golang:1.19 as buildenv
WORKDIR /dbx
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go mod download

FROM buildenv as build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go build ./...

FROM debian:bookworm as integration-test
RUN apt-get update && apt-get install -y postgresql wget procps gcc
RUN wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz && \
    rm go1.22.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin/
RUN echo 'local   all             all                                     trust' > /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'max_connections = 1000' >> /etc/postgresql/15/main/conf.d/connectionlimits.conf && \
    echo 'fsync = off' >> /etc/postgresql/15/main/conf.d/nosync.conf

RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v23.2.2.linux-amd64.tgz | tar xvz && \
    mv cockroach-v23.2.2.linux-amd64/cockroach /usr/local/bin/ && \
    mv cockroach-v23.2.2.linux-amd64/lib/* /usr/lib/
WORKDIR /dbx
ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    ./scripts/test-environment.sh go test ./...

FROM storjlabs/ci:slim as lint
WORKDIR /dbx
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    staticcheck ./...
#todo: we have too many violations
#RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
#    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
#    --mount=type=cache,target=/root/.cache/golangci-lint,id=golangcilint \
#    golangci-lint --config /go/ci/.golangci.yml -j=2 run
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go.min vet ./...
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    check-mod-tidy
RUN check-imports -race ./...
RUN check-downgrades
RUN check-copyright

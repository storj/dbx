# syntax=docker/dockerfile:1

FROM golang:1.22 as buildenv
WORKDIR /dbx
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go mod download

### integration-test

FROM debian:bookworm as integration-test
RUN apt-get update && apt-get install -y postgresql wget procps gcc
RUN wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz && \
    rm go1.22.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin
RUN echo 'local   all             all                                     trust' > /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/15/main/pg_hba.conf && \
    echo 'max_connections = 1000' >> /etc/postgresql/15/main/conf.d/connectionlimits.conf && \
    echo 'fsync = off' >> /etc/postgresql/15/main/conf.d/nosync.conf

RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v23.2.2.linux-amd64.tgz | tar xvz && \
    mv cockroach-v23.2.2.linux-amd64/cockroach /usr/local/bin/ && \
    mv cockroach-v23.2.2.linux-amd64/lib/* /usr/lib/


RUN apt-get update && apt-get install -y curl gpg
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
RUN apt-get update && apt-get install -y google-cloud-cli google-cloud-cli-spanner-emulator && \
    gcloud config configurations create emulator && \
    gcloud config set auth/disable_credentials true && \
    gcloud config set project storj-build && \
    gcloud config set api_endpoint_overrides/spanner http://localhost:9020/

RUN wget -qO- https://storage.googleapis.com/cloud-spanner-emulator/releases/1.5.18/cloud-spanner-emulator_linux_amd64-1.5.18.tar.gz | tar xvz && \
    chmod u+x gateway_main emulator_main && \
    cp gateway_main /usr/lib/google-cloud-sdk/bin/cloud_spanner_emulator/gateway_main && \
    cp emulator_main /usr/lib/google-cloud-sdk/bin/cloud_spanner_emulator/emulator_main

WORKDIR /dbx
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install golang.org/x/tools/cmd/bundle@latest
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go generate ./...
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    ./scripts/test-environment.sh go test ./...

### check-generated

FROM storjlabs/ci:slim as check-generated
WORKDIR /dbx
COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install golang.org/x/tools/cmd/bundle@latest
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go generate ./...
RUN git diff --exit-code

### lint

FROM storjlabs/ci:slim as lint
WORKDIR /dbx
COPY . .

RUN check-copyright

# this requries generated code, which is not ready for linting
RUN rm -rf testrun

RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    staticcheck ./...
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    --mount=type=cache,target=/root/.cache/golangci-lint,id=golangcilint \
    golangci-lint --config .golangci.yml -j=2 run
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    check-mod-tidy
RUN check-imports -race ./...
RUN check-downgrades

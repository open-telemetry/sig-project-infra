# syntax=docker/dockerfile:1.16@sha256:e2dd261f92e4b763d789984f6eab84be66ab4f5f08052316d8eb8f173593acf7
FROM golang:1.24.4-bullseye@sha256:6a4f10e74188b60ecc4c53f9fff88dbc220cc1ab982daace88c572ff9b627863 as builder

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ./otto && go build -o /out/otto ./cmd/otto

FROM debian:bullseye-slim@sha256:2f2307d7c75315ca7561e17a4e3aa95d58837f326954af08514044e8286e6d65
RUN useradd -m otto
COPY --from=builder /out/otto /usr/local/bin/otto
USER otto
WORKDIR /home/otto

# Expose the service port (default in main.go)
EXPOSE 8080

ENV OTTO_ADDR=:8080

ENTRYPOINT ["/usr/local/bin/otto"]

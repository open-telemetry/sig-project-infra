# syntax=docker/dockerfile:1.19@sha256:b6afd42430b15f2d2a4c5a02b919e98a525b785b1aaff16747d2f623364e39b6
FROM golang:1.24.6-bullseye@sha256:2cdc80dc25edcb96ada1654f73092f2928045d037581fa4aa7c40d18af7dd85a as builder

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ./otto && go build -o /out/otto ./cmd/otto

FROM debian:bullseye-slim@sha256:f807f4b16002c623115b0247dca6a55711c6b1ae821dc64fb8a2339e4ce2115d
RUN useradd -m otto
COPY --from=builder /out/otto /usr/local/bin/otto
USER otto
WORKDIR /home/otto

# Expose the service port (default in main.go)
EXPOSE 8080

ENV OTTO_ADDR=:8080

ENTRYPOINT ["/usr/local/bin/otto"]

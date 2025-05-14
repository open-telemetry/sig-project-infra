# syntax=docker/dockerfile:1.5@sha256:39b85bbfa7536a5feceb7372a0817649ecb2724562a38360f4d6a7782a409b14
FROM golang:1.24-bullseye@sha256:9cf2ca5977b21567c301d2528fc329b1ef1a6e391a50c48f4383feb9019b3a9c as builder

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ./otto && go build -o /out/otto ./cmd/otto

FROM debian:bullseye-slim@sha256:fdd75562fdcde1039c2480a1ea1cd2cf03b18b6e4cb551cabb03bde66ade8a5d
RUN useradd -m otto
COPY --from=builder /out/otto /usr/local/bin/otto
USER otto
WORKDIR /home/otto

# Expose the service port (default in main.go)
EXPOSE 8080

ENV OTTO_ADDR=:8080

ENTRYPOINT ["/usr/local/bin/otto"]

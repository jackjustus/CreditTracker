# syntax=docker/dockerfile:1

FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api-service ./cmd/api
FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM golang:1.26 AS migrate

# Build tags drop every driver but postgres; the default build pulls in
# clickhouse, sqlite, mssql and vertica.
RUN CGO_ENABLED=0 go install \
      -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb' \
      github.com/pressly/goose/v3/cmd/goose@v3.27.3

COPY db/migrations /migrations

ENV GOOSE_DRIVER=postgres
ENV GOOSE_MIGRATION_DIR=/migrations

ENTRYPOINT ["goose"]
CMD ["up"]

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=builder /api-service /api-service

EXPOSE 8080

USER nonroot:nonroot

CMD ["/api-service"]
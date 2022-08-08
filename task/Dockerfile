FROM golang:1.17-buster as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o application ./cmd/main.go

FROM alpine:3.15.4

ENV PORT=3000
ENV GRPC_AUTH="mts_teta_projects-auth-1:4000"
ENV GRPC_ANALYTIC="mts_teta_projects-analytic-1:4000"
ENV PROFILING=false
ENV PG_URL="postgres://team26:mNgd2ITbhVGd@91.185.93.23:5432/team26"
ENV JSON_DB_FILE="db.jsonl"
ENV KAFKA_URL="91.185.95.87:9094"
ENV KAFKA_ANALYTIC_TOPIC="team26-analytic"
ENV EMAIL_WORKERS=5
ENV EMAIL_RATE_LIMIT=3

COPY --from=builder /app/application /app/application
CMD ["/app/application"]

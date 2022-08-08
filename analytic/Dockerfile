FROM golang:1.17-buster as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
#RUN ls -la
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o application ./cmd/main.go

FROM alpine:3.15.4

ENV PORT=3000
ENV GRPC_AUTH_PORT=4000
ENV ACCESS_COOKIE="access_token"
ENV REFRESH_COOKIE="refresh_token"
ENV AUTHORIZATION_DB_CONNECTION_STRING="postgres://team26:mNgd2ITbhVGd@91.185.93.23:5432/team26"
ENV PROFILING=false
ENV KAFKA_URL="91.185.95.87:9094"
ENV KAFKA_ANALYTIC_TOPIC="team26-analytic"
ENV KAFKA_GROUP_ID="analytic"

COPY --from=builder /app/application /app/application
COPY *.yaml ./
CMD ["/app/application"]


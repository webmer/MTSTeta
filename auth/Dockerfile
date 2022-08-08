FROM golang:1.17-buster as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
#RUN ls -la
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o application ./cmd/main.go

FROM alpine:3.15.4

ENV PORT=3000
ENV GRPC_PORT=4000
ENV ACCESS_COOKIE="access_token"
ENV REFRESH_COOKIE="refresh_token"
ENV ACCESS_SECRET="team26_access_secret"
ENV REFRESH_SECRET="team26_refresh_secret"
ENV AUTHORIZATION_DB_CONNECTION_STRING="mongodb://team26:mNgd2ITbhVGd@91.185.93.34:27017/team26"
ENV AUTH_MONGO_MECH="SCRAM-SHA-1"
ENV MONGO_DB_NAME="team26"
ENV MONGO_USER_NAME="team26"
ENV MONGO_USER_PASS="mNgd2ITbhVGd"
ENV PROFILING=false

COPY --from=builder /app/application /app/application
COPY *.yaml ./
CMD ["/app/application", "-c", "/app/config.yaml"]


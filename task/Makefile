start:
	go run cmd/main.go

cover:
	go test -tags=all -coverprofile=coverage.out ./...

coverage:
	go tool cover -func=coverage.out

unit_tests:
	go test -tags=unit ./...

integration_tests:
	go test -tags=integration ./...

gen_swagger:
	swag init --parseDependency --parseInternal -g internal/adapters/http/task.go

gen_proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/auth.proto

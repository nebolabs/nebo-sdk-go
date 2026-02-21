.PHONY: proto build test

proto:
	protoc \
		--go_out=pb --go_opt=module=github.com/neboloop/nebo-sdk-go/pb \
		--go-grpc_out=pb --go-grpc_opt=module=github.com/neboloop/nebo-sdk-go/pb \
		-I. \
		proto/apps/v0/*.proto

build:
	go build ./...

test:
	go test ./...

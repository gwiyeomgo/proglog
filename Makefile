compile1:
	protoc api/v1/log.proto --go_out=. --go_opt=paths=source_relative --proto_path=.

compile2:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/v1/log.proto

test:
	go test -race ./...
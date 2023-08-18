compile:
	protoc structureDataWithProtobuf/api/v1/log.proto --go_out=. --go_opt=paths=source_relative --proto_path=.
test:
	go test -race ./...
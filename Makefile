run:
	go run main.go

proto:
	rm -rf protobuf/protogen/*.go
	protoc --proto_path=protobuf/proto --go_out=protobuf/protogen --go_opt=paths=source_relative \
	--go-grpc_out=protobuf/protogen --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=protobuf/protogen --grpc-gateway_opt=paths=source_relative \
	./protobuf/proto/*.proto 

.PHONY: proto run
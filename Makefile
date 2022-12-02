proto: protoc_install proto_core

protoc_install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

GOPATH:=$(shell go env GOPATH)
PROTOC:=protoc
PROTOC:=$(PROTOC) --plugin=protoc-gen-go-grpc=$(GOPATH)/bin/protoc-gen-go-grpc
PROTOC:=$(PROTOC) --plugin=protoc-gen-go=$(GOPATH)/bin/protoc-gen-go
PROTOC:=$(PROTOC) --go_opt=module=github.com/d-ion/isglb --go_out=.
PROTOC:=$(PROTOC) --go-grpc_opt=module=github.com/d-ion/isglb --go-grpc_out=.
PROTOC:=$(PROTOC) -I ./

proto_core:
	$(PROTOC) proto/isglb.proto
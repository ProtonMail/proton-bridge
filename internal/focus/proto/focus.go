// Package proto provides the gRPC definition of the focus service.
package proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative focus.proto

module github.com/erda-project/erda/cmd/ai-proxy/integration-tests

go 1.22

toolchain go1.24.5

require (
	github.com/erda-project/erda-proto-go v1.4.0
	github.com/sashabaranov/go-openai v1.40.1
)

require (
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/erda-project/erda-infra v1.0.8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/sys v0.0.0-20220429233432-b5fbb4746d32 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210820002220-43fce44e7af1 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/erda-project/erda-proto-go v1.4.0 => ../../../api/proto-go

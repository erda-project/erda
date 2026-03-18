module github.com/erda-project/erda/cmd/ai-proxy/integration-tests

go 1.24.0

toolchain go1.24.5

require (
	github.com/sashabaranov/go-openai v1.40.1
	github.com/stretchr/testify v1.7.5
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/erda-project/erda-proto-go v1.4.0 => ../../../api/proto-go

module github.com/liliang-cn/ohman

go 1.24.0

toolchain go1.24.0

// Package aliases map module paths to package import paths
package github.com/liliang-cn/ohman/cmd/ohman => github.com/liliang-cn/ohman

require (
	github.com/openai/openai-go/v3 v3.15.0
	github.com/spf13/cobra v1.10.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)

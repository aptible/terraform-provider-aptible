//go:build tools
// +build tools

package tools

import (
	_ "github.com/bflad/tfproviderdocs"
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/katbyte/terrafmt"
)

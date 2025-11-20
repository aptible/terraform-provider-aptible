//go:build tools
// +build tools

package tools

import (
	_ "github.com/bflad/tfproviderdocs"
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
	_ "github.com/katbyte/terrafmt"
)

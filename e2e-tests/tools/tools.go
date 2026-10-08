//go:build tools
// +build tools

package tools

import (
	_ "actionlint.kjanat.dev/cmd/actionlint"
	_ "github.com/fullstorydev/grpcurl/cmd/grpcurl"
	_ "github.com/mcpchecker/mcpchecker/cmd/mcpchecker"
)

//go:build tools

// Package tools pins the versions of CI-only security scanners (gosec,
// govulncheck) so dependabot can bump them like any other dependency. The
// `tools` build tag means this file is never compiled into anything — the
// blank imports exist solely to keep the corresponding `require` lines
// alive through `go mod tidy`.
package tools

import (
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "golang.org/x/vuln/cmd/govulncheck"
)

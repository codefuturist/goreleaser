// SOURCE: github.com/goreleaser/goreleaser-pro/v2 @ v2.14.0
// FILE:   pkg/config/matrix.go
// COPIED: 2026-02-25 (1:1 copy, do not manually edit)
// UPDATE: Replace this file from the latest goreleaser-pro release:
//   git clone https://github.com/goreleaser/goreleaser-pro.git /tmp/gp
//   cp /tmp/gp/pkg/config/matrix.go pkg/config/matrix.go
//   go build ./... && go test ./pkg/config/... ./internal/pipe/...
package config

// only in goreleaser pro
type Matrix map[string][]string

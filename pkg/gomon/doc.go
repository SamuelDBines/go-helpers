// Package gomon loads job definitions from gomon.json / gomon.yaml / gomon.toml,
// matches watch paths with simple glob rules, and runs shell commands with optional
// live reload. Use [LoadConfig] and [ApplyDefaults] to read configuration, [RunJob]
// to execute a job once, and [RunWatch] to restart when watched files change.
//
// The CLI binary is published from repository github.com/SamuelDBines/gomon (cmd/gomon).
// This tree under go-helpers mirrors that package so importers can use
// github.com/SamuelDBines/go-helpers/pkg/gomon without a second module; keep it
// aligned when fixing bugs in the standalone repo.
//
// Tests: go test ./pkg/gomon/...
package gomon

//go:build packed

package main

// Production builds embed config, templates and static files (see `make build`).
// Dev runs read them from disk, because GoFrame prefers packed resources over files.
import _ "alpaca-app/internal/packed"

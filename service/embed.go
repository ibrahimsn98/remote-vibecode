// Package embed provides embedded web assets for the remote-vibecode service
package embed

import (
	"embed"
)

//go:embed web/*
var FS embed.FS

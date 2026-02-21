package app

import "embed"

//go:embed spec/openapi.yaml
var specFS embed.FS

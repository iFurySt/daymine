package webassets

import "embed"

// Files contains the production frontend bundle. During development, Vite writes
// into dist; the checked-in placeholder keeps go:embed valid before the first build.
//
//go:embed all:dist
var Files embed.FS

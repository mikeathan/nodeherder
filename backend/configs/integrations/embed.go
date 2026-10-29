// Package integrationdefaults embeds reviewed provider seeds in binary-only installations.
package integrationdefaults

import "embed"

// Files supplies missing defaults; runtime host definitions are never overwritten.
//
//go:embed waste_collection/*.yaml weather/*.yaml
var Files embed.FS

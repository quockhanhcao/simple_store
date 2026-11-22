package web

import "embed"

// StoreFS holds web/store/** inside the binary. //go:embed can't reach "..", which is why
// this file lives next to the folder. "all:" also includes files starting with . or _.
//
//go:embed all:store
var StoreFS embed.FS

// Command pack bundles config, templates and static files into internal/packed,
// using GoFrame's gres (same result as `gf pack`, without installing the CLI).
// Run from the repo root: go run ./hack/pack
package main

import (
	"log"

	"github.com/gogf/gf/v2/os/gres"
)

func main() {
	err := gres.PackToGoFileWithOption(
		"manifest/config,resource",
		"internal/packed/packed.go",
		"packed",
		gres.Option{KeepPath: true},
	)
	if err != nil {
		log.Fatal(err)
	}
}

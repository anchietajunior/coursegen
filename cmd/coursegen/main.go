// Command coursegen orchestrates the production of course lessons with AI
// agents — one lesson per isolated session, sequentially, with a clean context
// between lessons and a minimal context pack to economize tokens.
package main

import (
	"os"

	"github.com/anchietajunior/coursegen/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}

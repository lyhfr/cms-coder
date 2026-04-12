package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"cmscoder-web-server/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}

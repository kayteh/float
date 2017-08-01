package main

import (
	"github.com/kayteh/float/shim/fn"
)

func main() {
	fn.Handle(func(ctx *fn.FuncCtx) {

		// ctx.Response.StatusCode = ?
		ctx.WriteString("hello world!")
		ctx.SetHeader("X-Test", "tseT-X")

	})
}

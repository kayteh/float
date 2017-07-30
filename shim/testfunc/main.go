package main

import (
	"encoding/json"

	"github.com/kayteh/float/shim/fn"
)

func main() {
	fn.Handle(func(ctx *fn.Context) {

		ctx.Response.StatusCode = 418
		json.NewEncoder(ctx).Encode(map[string]string{
			"hello": "world",
		})

	})
}

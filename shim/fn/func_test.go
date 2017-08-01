package fn

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func ExampleFuncCtx_WriteJSON() {
	Handle(func(ctx *FuncCtx) {

		type exampleType struct {
			Hello string `json:"hello"`
			World string `json:"world"`
		}

		ctx.WriteJSON(&exampleType{
			Hello: "こんばんわ",
			World: "みんなさん",
		})

	})
}

func TestHandle(t *testing.T) {
	rr := &bytes.Buffer{}
	ww := &bytes.Buffer{}

	err := json.NewEncoder(rr).Encode(&Request{
		Body: "hello!",
	})
	if err != nil {
		t.Error(err)
		return
	}

	FHandle(rr, ww, os.Stderr, func(ctx *FuncCtx) {

		if ctx.Request.Body != "hello!" {
			t.Error("stdin was not parsed correctly", ctx)
		}

		ctx.SetHeader("X-Test", "tseT-X")
		err := json.NewEncoder(ctx).Encode(map[string]string{
			"hi": "how are you",
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	var rsp Response
	err = json.NewDecoder(ww).Decode(&rsp)
	if err != nil {
		t.Error(err)
		return
	}

	body := bytes.NewBufferString(rsp.Body)

	var out map[string]string
	err = json.NewDecoder(body).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	if out["hi"] != "how are you" {
		t.Error("EVERYTHING IS WRONG", out)
	}
}

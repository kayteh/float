package main

import (
	"os"
)

func main() {
	os.Stdout.WriteString(`{"status_code":404, "headers": {"X-Test": "tseT-X"}, "body":"hello world!"}`)
	os.Stderr.WriteString("called\n")
}

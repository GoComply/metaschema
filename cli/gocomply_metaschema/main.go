package main

import (
	"github.com/gocomply/metaschema/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
